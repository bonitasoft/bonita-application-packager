package main

import (
	"archive/zip"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
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
)

const (
	_sep                    = os.PathSeparator
	dockerImagePrefix       = "bonita-application-"
	defaultBaseImageName    = "bonita"
	defaultBaseImageNameSp  = "quay.io/bonitasoft/bonita-subscription"
	defaultBaseImageVersion = "latest"
)

var (
	// Version is overridden via ldflags
	Version = "development"

	// Flags:
	tomcatFlag = flag.Bool("tomcat", false, `Choose to build a Bonita Tomcat bundle containing your application
use -bonita-tomcat-bundle to specify the path to the Bonita tomcat bundle file (Bonita*.zip); otherwise the file is looked for in the current folder`)
	dockerFlag = flag.Bool("docker", false, fmt.Sprintf(
		`Choose to build a docker image containing your application,
use -tag to specify the name of your built image
By default, it builds a 'Community' Docker image
use -subscription to build a 'Subscription' Docker image (you must have the rights to download Bonita Subscription Docker base image from Bonita Artifact Repository)
use -base-image-name to specify a Bonita docker base image different from the default, which is
    '%s' in Community edition
    '%s' in Subscription edition
use -base-image-version to specify a Bonita docker base image version different from the default ('%s')
use -registry-username and -registry-password if you need to authenticate against the docker image registry to pull Bonita docker base image`,
		defaultBaseImageName, defaultBaseImageNameSp, defaultBaseImageVersion))
	dockerSubscription = flag.Bool("subscription", false, "Choose to build a Subscription-based docker image (default build a Community image)")
	tag                = flag.String("tag", dockerImagePrefix, "Docker image tag to use when building")
	verbose            = flag.Bool("verbose", false, "Enable verbose (debug) mode")
	baseImageName      = flag.String("base-image-name", "", "Specify Bonita base docker image name")
	baseImageVersion   = flag.String("base-image-version", "", "Specify Bonita base docker image version")
	registryUsername   = flag.String("registry-username", "", "Specify username to authenticate against Bonita base docker image Registry")
	registryPassword   = flag.String("registry-password", "", "Specify corresponding password to authenticate against Bonita base docker image Registry")
	configurationFile  = flag.String("configuration-file", "", "(Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)")
	tomcatBundleFile   = flag.String("bonita-tomcat-bundle", "", "(Optional) Specify path to the Bonita tomcat bundle file (Bonita*.zip) used to build")
	version            = flag.Bool("version", false, "Print tool version and exit")

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
		fmt.Fprintf(w, "\t%s -verbose -docker -subscription -registry-username customer@bonitasoft.com -registry-password <MY_SECRET_PWD> /tmp/my-application.zip\n", os.Args[0])
		fmt.Fprintf(w, "\t%s -verbose -tomcat /tmp/my-application.zip\n", os.Args[0])
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
	if *tomcatBundleFile != "" {
		fmt.Println("Bonita Tomcat bundle file :", *tomcatBundleFile)
	}
	if *configurationFile != "" {
		fmt.Println("Bonita configuration file :", *configurationFile)
	}
	var dockerEdition = "community"
	if *dockerFlag {
		fmt.Println("Docker image tag name     :", *tag)
		if *dockerSubscription {
			dockerEdition = "subscription"
		}
		fmt.Println("Docker image edition      :", dockerEdition)
	}

	if *tomcatFlag {
		buildTomcatBundle()
	}

	if *dockerFlag {
		buildDockerImage(dockerEdition)
	}
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

func buildDockerImage(dockerEdition string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = imageBuild(cli, dockerEdition)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func imageBuild(dockerClient *client.Client, edition string) error {
	fmt.Printf("Building %s Docker image\n", edition)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	if *baseImageName == "" {
		if *dockerSubscription {
			*baseImageName = defaultBaseImageNameSp
		} else {
			*baseImageName = defaultBaseImageName
		}
	}
	if *baseImageVersion == "" {
		*baseImageVersion = defaultBaseImageVersion
	}
	if err := pullBaseImage(*baseImageName, *baseImageVersion, dockerClient, ctx); err != nil {
		return err
	}

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

	fullDockerImageName := *tag
	if *tag == dockerImagePrefix {
		fullDockerImageName = *tag + edition
	}
	if err := buildCustomDockerImage(baseImageName, baseImageVersion, ctx, dockerClient, dockerContext, fullDockerImageName); err != nil {
		return err
	}
	fmt.Printf("\nSuccessfully created Docker image '%s'\n\n", fullDockerImageName)
	return nil
}

func pullBaseImage(baseImageName, baseImageVersion string, dockerClient *client.Client, ctx context.Context) error {
	authConfig := types.AuthConfig{
		Username: *registryUsername,
		Password: *registryPassword,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	if *verbose {
		fmt.Println("Pulling base docker image: " + baseImageName + ":" + baseImageVersion)
	}

	out, err := dockerClient.ImagePull(ctx, baseImageName+":"+baseImageVersion, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer out.Close()

	if *verbose {
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		err = jsonmessage.DisplayJSONMessagesStream(out, os.Stderr, termFd, isTerm, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildCustomDockerImage(baseImageName *string, baseImageVersion *string, ctx context.Context, dockerClient *client.Client, dockerContext io.ReadCloser, fullDockerImageName string) error {
	dockerfile := "Dockerfile"
	opts := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{fullDockerImageName},
		Remove:     true,
		BuildArgs: map[string]*string{
			"BONITA_IMAGE_NAME":    baseImageName,
			"BONITA_IMAGE_VERSION": baseImageVersion},
	}
	if *verbose {
		fmt.Println("Using base docker image: " + *baseImageName + ":" + *baseImageVersion)
		fmt.Println("Building new image: " + fullDockerImageName)
	}

	res, err := dockerClient.ImageBuild(ctx, dockerContext, opts)
	if err != nil {
		fmt.Println("Error building image")
		return err
	}

	defer res.Body.Close()

	if *verbose {
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		err = jsonmessage.DisplayJSONMessagesStream(res.Body, os.Stderr, termFd, isTerm, nil)
		if err != nil {
			return err
		}
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
