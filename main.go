package main

import (
	"archive/zip"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	cp "github.com/otiai10/copy"
	t "golang.org/x/term"
)

const (
	_sep             = os.PathSeparator
	defaultImageTag  = "my-bonita-application:latest"
	defaultBaseImage = "bonita:latest"
)

var (
	// Version is overridden via ldflags
	Version = "development"

	// Flags:
	// - Common usage:
	version           = flag.Bool("version", false, "Print tool version and exit")
	verbose           = flag.Bool("verbose", false, "Enable verbose (debug) mode")
	configurationFile = flag.String("configuration-file", "", "(Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)")

	// - Tomcat usage:
	tomcatFlag = flag.Bool("tomcat", false, `Choose to build a Bonita Tomcat bundle containing your application
use -bonita-tomcat-bundle to specify the path to the Bonita tomcat bundle file (Bonita*.zip); otherwise the file is looked for in the current folder`)
	tomcatBundleFile = flag.String("bonita-tomcat-bundle", "", "(Optional) Specify path to the Bonita tomcat bundle file (Bonita*.zip) used to build")

	// - Docker usage:
	dockerFlag = flag.Bool("docker", false, fmt.Sprintf(
		`Choose to build a docker image containing your application,
use -tag to specify the name of your built image
use -bonita-base-image to specify a Bonita docker base image different from the default, which is '%s'
use -registry-username and -registry-password if you need to authenticate against the docker image registry to pull Bonita docker base image`,
		defaultBaseImage))
	tag              = flag.String("tag", defaultImageTag, "Docker image tag to use when building")
	baseImage        = flag.String("bonita-base-image", defaultBaseImage, "Specify Bonita base docker image")
	registryUsername = flag.String("registry-username", "", "Specify username to authenticate against Bonita base docker image Registry")
	registryPassword = flag.String("registry-password", "", `Specify password to authenticate against Bonita base docker image Registry
If -registry-username is provided and not -registry-password, password will be prompted interactively and never issued to the console`)

	// First argument of the command line pointing to the custom application zip file
	appPath string

	// Go directive to include Dockerfile in the binary:
	//go:embed Dockerfile
	dockFile []byte
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Bonita Application Packager version %s\n", Version)
		fmt.Fprintln(w, "This tool allows to build a Bonita Tomcat bundle or a Bonita Docker image containing your custom application.")
		fmt.Fprintf(w, "%s [-tomcat|-docker] [OPTIONS] PATH_TO_APPLICATION_ZIP_FILE\n", os.Args[0])
		fmt.Fprintf(w, "Options are:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(w, "\nExample:\n")
		fmt.Fprintf(w, "\t%s -docker -bonita-base-image bonita:8.0.0 -tag my-bonita-application:1.0.0 /path/to/my-custom-application.zip\n", os.Args[0])
		fmt.Fprintf(w, "\t%s -tomcat -bonita-tomcat-bundle /path/to/BonitaCommunity-2023.1-u0.zip /path/to/my-custom-application.zip\n", os.Args[0])
	}
	flag.Parse()

	if *version {
		fmt.Printf("Bonita Application Packager version %s\n", Version)
		os.Exit(0)
	}

	if !*dockerFlag && !*tomcatFlag {
		ExitWithError("Please specify '-tomcat' if you want to build a Bonita Tomcat Bundle or '-docker' if you want to build a Bonita Docker image.")
	}

	arguments := flag.Args()
	if len(arguments) != 1 {
		ExitWithError("Please provide one and only one argument that points to your application ZIP file.")
	}

	appPath = arguments[0]
	if !strings.HasSuffix(appPath, ".zip") {
		ExitWithError("Application file '%s' is not a ZIP file.", appPath)
	}
	if !Exists(appPath) {
		ExitWithError("Application ZIP file '%s' does not exist.", appPath)
	}

	if *configurationFile != "" {
		if !strings.HasSuffix(*configurationFile, ".bconf") {
			ExitWithError("Bonita configuration file '%s' is not a .bconf file.", *configurationFile)
		}
		if !Exists(*configurationFile) {
			ExitWithError("Bonita configuration file '%s' does not exist.", *configurationFile)
		}
	}

	fmt.Println("Verbose mode              :", *verbose)
	fmt.Println("Build Tomcat bundle       :", *tomcatFlag)
	fmt.Println("Build Docker image        :", *dockerFlag)
	fmt.Println("Custom application        :", appPath)
	if *configurationFile != "" {
		fmt.Println("Bonita configuration file :", *configurationFile)
	}
	if *tomcatBundleFile != "" {
		fmt.Println("Bonita Tomcat bundle file :", *tomcatBundleFile)
	}
	if *dockerFlag {
		fmt.Println("Bonita Docker base image  :", *baseImage)
		fmt.Println("Docker image tag name     :", *tag)
	}

	if *tomcatFlag {
		buildTomcatBundle()
	}

	if *dockerFlag {
		buildDockerImage()
	}
}

func printFinalNote(additionalNote string) {
	fmt.Println("\nNOTE: if your custom application is using pages from Bonita Admin or User applications, " +
		additionalNote + " in order to install those pages, else, your application will fail at install.")
}

func ExitWithError(message string, messageArgs ...any) {
	fmt.Printf("\nError: "+message+"\n\n", messageArgs...)
	flag.Usage()
	os.Exit(1)
}

func buildTomcatBundle() {
	bundleAbsolutePath := validateTomcatBundle()
	if bundleAbsolutePath == "" {
		return
	}
	if Exists("output") {
		if *verbose {
			fmt.Println("Cleaning 'output/' folder")
		}
		if err := os.RemoveAll("output"); err != nil {
			fmt.Println("Failed to clean 'output/' folder")
			panic(err)
		}
	}
	bundleNameAndPath := filepath.Base(bundleAbsolutePath)                      // just the name of the zip file without the path
	bundleName := bundleNameAndPath[0:strings.Index(bundleNameAndPath, ".zip")] // just the name without '.zip'
	fmt.Printf("Unpacking Bonita Tomcat bundle %s.zip\n", bundleName)
	unzipFile(bundleAbsolutePath, "output")
	fmt.Println("Unpacking Bonita WAR file")
	unzipFile(filepath.Join("output", bundleName, "server", "webapps", "bonita.war"), filepath.Join("output", bundleName, "server", "webapps", "bonita"))
	if *verbose {
		fmt.Println("Removing unpacked Bonita WAR file")
	}
	if err := os.Remove(filepath.Join("output", bundleName, "server", "webapps", "bonita.war")); err != nil {
		panic(err)
	}
	fmt.Println("Copying your custom application inside Bonita")
	copyResourceToCustomAppFolder(bundleName, appPath)
	if *configurationFile != "" {
		fmt.Println("Copying your Bonita configuration file inside Bonita")
		copyResourceToCustomAppFolder(bundleName, *configurationFile)
	}
	fmt.Println("Re-packing Bonita bundle containing your application")
	err := zipDirectory(filepath.Join("output", bundleName+"-application.zip"), filepath.Join("output", bundleName), bundleName)
	if err != nil {
		panic(err)
	}
	tempfolderToZip := filepath.Join("output", bundleName)
	if Exists(tempfolderToZip) {
		if *verbose {
			fmt.Println("Cleaning temporary folder structure")
		}
		if err := os.RemoveAll(tempfolderToZip); err != nil {
			panic(err)
		}
	}
	fmt.Println("\nSuccessfully re-packaged self-contained application:", filepath.Join("output", bundleName+"-application.zip"))
	fmt.Println("\nTo use it, simply unzip it like your usual Bonita Tomcat bundle, and run ./start-bonita[.sh|.bat]")
	fmt.Println("More info at https://documentation.bonitasoft.com/bonita/latest/runtime/tomcat-bundle")

	printFinalNote("ensure to set the Bonita runtime property 'bonita.runtime.custom-application.install-provided-pages=true' in bundle configuration")
}

// check if the Bonita Tomcat bundle is passed as parameter, or found in current folder, exists
func validateTomcatBundle() string {
	if *tomcatBundleFile != "" {
		if !Exists(*tomcatBundleFile) {
			fmt.Println("Bonita Tomcat bundle file passed as parameter does not exist: " + *tomcatBundleFile)
			return ""
		} else if !strings.HasSuffix(*tomcatBundleFile, ".zip") {
			fmt.Println("Bonita Tomcat bundle file passed as parameter is not a proper Bonita Tomcat bundle ZIP file: " + *tomcatBundleFile)
			return ""
		} else {
			if *verbose {
				fmt.Println("Using Bonita Tomcat bundle file passed as parameter", *tomcatBundleFile)
			}
			return *tomcatBundleFile
		}
	} else {
		// Try to find a Bonita zip file in current folder:
		bundleMatches, err := filepath.Glob("Bonita*.zip")
		if err != nil {
			panic(err)
		}
		if len(bundleMatches) == 0 {
			fmt.Println("Bonita Tomcat Bundle not found in current folder.")
			fmt.Println("Please copy it here (Eg. BonitaCommunity-2023.1-u0.zip, BonitaSubscription-2023.1-u2.zip)")
			fmt.Println("or use parameter -bonita-tomcat-bundle <PATH_TO_TOMCAT_BUNDLE> if stored somewhere else.")
			fmt.Println("Then re-run this program")
			return ""
		}
		if *verbose {
			fmt.Println("Using Bonita Tomcat bundle file found in current folder", bundleMatches[0])
		}
		return bundleMatches[0]
	}
}

func copyResourceToCustomAppFolder(bundleName string, resource string) {
	err := cp.Copy(resource, filepath.Join("output", bundleName, "server", "webapps", "bonita", "WEB-INF", "classes", "my-application", filepath.Base(resource)))
	if err != nil {
		panic(err)
	}
}

func buildDockerImage() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = imageBuild(cli)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	printFinalNote("ensure to set the environment variable 'INSTALL_PROVIDED_PAGES=true' when running container")
}

func imageBuild(dockerClient *client.Client) error {
	fmt.Println("Building Docker image")
	var timeout = printMsgIfVerbose("Using docker context timeout", time.Minute*5)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// create temporary folder to store the docker context needed to build the image
	dockerContextDir, err := os.MkdirTemp("", "docker-context")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerContextDir)

	dockerResourcesDir := filepath.Join(dockerContextDir, "resources")

	// copy application file in Docker build context resources folder:
	appName := filepath.Base(appPath)
	if err := cp.Copy(appPath, filepath.Join(dockerResourcesDir, appName)); err != nil {
		return err
	}

	bconfName := ""
	if *configurationFile != "" {
		// copy bconf file in Docker build context resources folder:
		bconfName = filepath.Base(*configurationFile)
		if err := cp.Copy(*configurationFile, filepath.Join(dockerResourcesDir, bconfName)); err != nil {
			return err
		}
	}

	dockerfile := "Dockerfile"
	// write the Dockerfile embedded in this program to Docker build context folder:
	if err := os.WriteFile(filepath.Join(dockerContextDir, dockerfile), dockFile, 0644); err != nil {
		return err
	}

	// On Windows `archive.TarWithOptions` requires an absolute directory path
	dockerContextDirAbsPath, err := filepath.Abs(dockerContextDir)
	if err != nil {
		return err
	}

	dockerContext, err := archive.TarWithOptions(dockerContextDirAbsPath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer dockerContext.Close()

	if err := buildCustomDockerImage(ctx, dockerClient, dockerContext); err != nil {
		return err
	}
	fmt.Printf("\nSuccessfully created Docker image '%s'\n\n", *tag)
	fmt.Printf("To use it, run appropriate command:\n")
	fmt.Printf("- Community release    : docker run --name my-bonita-app -d -p 8080:8080 %s\n", *tag)
	fmt.Printf("- Subscription release : docker run --name my-bonita-app -h <hostname> -v <license-folder>:/opt/bonita_lic/ -d -p 8080:8080 %s\n", *tag)
	fmt.Printf("Read https://documentation.bonitasoft.com/bonita/latest/runtime/bonita-docker-installation for complete options on how to run a Bonita-based Docker container.\n")
	return nil
}

func buildCustomDockerImage(ctx context.Context, dockerClient *client.Client, dockerContext io.ReadCloser) error {
	dockerfile := "Dockerfile"
	opts := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{*tag},
		Remove:     true,
		BuildArgs: map[string]*string{
			"BONITA_BASE_IMAGE": baseImage},
	}
	if *verbose {
		fmt.Println("Building new image:", *tag)
	}

	if *registryUsername != "" && *registryPassword == "" {
		fmt.Printf("Enter your password to access the Docker Registry corresponding to '%v':", *registryUsername)
		p, err := t.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("Error reading your password")
			return err
		}
		*registryPassword = string(p)
		fmt.Println() // to make the next print on a fresh new line
	}

	// configure registry authentication
	if *registryUsername != "" && *registryPassword != "" {
		registryName, _, found := strings.Cut(*baseImage, "/")
		if !found {
			// if no registry found, set default docker registry
			registryName = "docker.io"
		}
		if *verbose {
			fmt.Println("Authenticating to registry:", registryName)
		}
		opts.AuthConfigs = map[string]types.AuthConfig{
			registryName: {
				Username: *registryUsername,
				Password: *registryPassword,
			},
		}
	}

	res, err := dockerClient.ImageBuild(ctx, dockerContext, opts)
	if err != nil {
		fmt.Println("Error building image")
		return err
	}

	defer res.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	err = jsonmessage.DisplayJSONMessagesStream(res.Body, os.Stderr, termFd, isTerm, nil)
	if err != nil {
		return err
	}
	return nil
}

func unzipFile(zipFile string, outputDir string) {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(outputDir, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func zipDirectory(zipFilename string, baseDir string, baseInZip string) error {
	outFile, err := os.Create(zipFilename)
	if err != nil {
		return err
	}

	w := zip.NewWriter(outFile)

	if err := addFilesToZip(w, baseDir, baseInZip); err != nil {
		_ = outFile.Close()
		return err
	}

	if err := w.Close(); err != nil {
		_ = outFile.Close()
		return errors.New("Warning: closing zipfile writer failed: " + err.Error())
	}

	if err := outFile.Close(); err != nil {
		return errors.New("Warning: closing zipfile failed: " + err.Error())
	}

	return nil
}

func addFilesToZip(w *zip.Writer, basePath, baseInZip string) error {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullfilepath := filepath.Join(basePath, file.Name())
		if _, err := os.Stat(fullfilepath); os.IsNotExist(err) {
			// ensure the file exists. For example a symlink pointing to a non-existing location might be listed but not actually exist
			continue
		}

		if file.Mode()&os.ModeSymlink != 0 {
			// ignore symlinks alltogether
			continue
		}

		if file.IsDir() {
			// create dir first
			path := filepath.Join(baseInZip, file.Name())
			_, err := w.Create(path + "/")
			if err != nil {
				return err
			}
			// then add files inside it
			if err := addFilesToZip(w, fullfilepath, baseInZip+"/"+file.Name()); err != nil {
				return err
			}
		} else if file.Mode().IsRegular() {
			dat, err := ioutil.ReadFile(fullfilepath)
			if err != nil {
				return err
			}
			fh := &zip.FileHeader{Name: baseInZip + "/" + file.Name()}
			fh.SetMode(file.Mode())
			f, err := w.CreateHeader(fh)
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		} else {
			// we ignore non-regular files because they are scary
		}
	}
	return nil
}

func printMsgIfVerbose[T any](message string, o T) T {
	if *verbose {
		fmt.Println(message, o)
	}
	return o
}
