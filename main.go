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
)

const (
	_sep              = os.PathSeparator
	dockerImagePrefix = "bonita-application-"
)

var (
	// Flags:
	tomcatFlag = flag.Bool("tomcat", false, "Choose to build a Tomcat bundle containing your application")
	dockerFlag = flag.Bool("docker", false, `Choose to build a docker image containing your application,
	use -tag to specify the name of your built image
	By default, it builds a 'Community' Docker image
	use -subscription to build a 'Subscription' Docker image (you must have the rights to download Bonita Subscription Docker base image from Bonita Artifact Repository)`)
	dockerSubscription = flag.Bool("subscription", false, "Choose to build a Subscription-based docker image (default build a Community image)")
	tag                = flag.String("tag", dockerImagePrefix, "Docker image tag to use when building")
	verbose            = flag.Bool("verbose", false, "Enable verbose (debug) mode")
	appPath            string

	// Go directive to include Dockerfile in the binary:
	//go:embed Dockerfile
	dockFile []byte
)

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "This tool allows to build a Bonita Tomcat bundle or a Bonita Docker image containing your custom application")
		fmt.Fprintf(w, "%s [-tomcat|-docker] [OPTIONS] PATH_TO_APPLICATION_ZIP_FILE\n", os.Args[0])
		fmt.Fprintf(w, "Options are:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if !*dockerFlag && !*tomcatFlag {
		fmt.Printf("Please specify '-tomcat' if you want to build a Bonita Tomcat Bundle or '-docker' if you want to build a Bonita Docker image\n\n")
		flag.Usage()
		os.Exit(1)
	}

	arguments := flag.Args()
	// fmt.Println("flag.Args():", arguments)
	if len(arguments) != 1 {
		fmt.Printf("Please provide one and only one argument that points to your application ZIP file\n\n")
		flag.Usage()
		os.Exit(1)
	}

	appPath = arguments[0]
	if !Exists(appPath) {
		fmt.Printf("Application ZIP file %s does not exist\n\n", appPath)
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Verbose mode          :", *verbose)
	fmt.Println("Build Tomcat bundle   :", *tomcatFlag)
	fmt.Println("Build Docker image    :", *dockerFlag)
	fmt.Println("Custom application    :", appPath)
	var dockerEdition = "community"
	if *dockerFlag {
		fmt.Println("Docker image tag name :", *tag)
		if *dockerSubscription {
			dockerEdition = "subscription"
		}
		fmt.Println("Docker image edition  :", dockerEdition)
	}

	if *tomcatFlag {
		buildTomcatBundle()
	}

	if *dockerFlag {
		buildDockerImage(dockerEdition)
	}
}

func buildTomcatBundle() {
	// Try to find a Bonita zip file in current folder:
	bundleMatches, err := filepath.Glob("Bonita*.zip")
	if err != nil {
		panic(err)
	}
	if len(bundleMatches) == 0 {
		fmt.Println("Bonita Tomcat Bundle not found in current folder")
		fmt.Println("Please copy it here (Eg. BonitaCommunity-2023.1-u0.zip, BonitaSubscription-2023.1-u2.zip)")
		fmt.Println("and then re-run this program")
		return
	}
	if Exists("output") {
		if *verbose {
			fmt.Println("Cleaning 'output' directory")
		}
		if err := os.RemoveAll("output"); err != nil {
			panic(err)
		}
	}
	bundleNameAndPath := bundleMatches[0]
	bundleName := bundleNameAndPath[0:strings.Index(bundleNameAndPath, ".zip")] // until end of string
	fmt.Printf("Unpacking Bonita Tomcat bundle %s.zip\n", bundleName)
	unzipFile(bundleNameAndPath, "output")
	fmt.Println("Unpacking Bonita WAR file")
	unzipFile(filepath.Join("output", bundleName, "server", "webapps", "bonita.war"), filepath.Join("output", bundleName, "server", "webapps", "bonita"))
	if *verbose {
		fmt.Println("Removing unpacked Bonita WAR file")
	}
	if err := os.Remove(filepath.Join("output", bundleName, "server", "webapps", "bonita.war")); err != nil {
		panic(err)
	}
	fmt.Println("Copying your custom application inside Bonita")
	err = cp.Copy(appPath, filepath.Join("output", bundleName, "server", "webapps", "bonita", "WEB-INF", "classes", "my-application", filepath.Base(appPath)))
	if err != nil {
		panic(err)
	}
	fmt.Println("Re-packing Bonita bundle containing your application")
	err = zipDirectory(filepath.Join("output", bundleName+"-application.zip"), filepath.Join("output", bundleName), bundleName)
	if err != nil {
		panic(err)
	}
	tempfolderToZip := filepath.Join("output", bundleName)
	if Exists(tempfolderToZip) {
		if *verbose {
			fmt.Println("Cleaning temporary directory structure")
		}
		if err := os.RemoveAll(tempfolderToZip); err != nil {
			panic(err)
		}
	}
	fmt.Println("\nSuccessfully re-packaged self-contained application:", filepath.Join("output", bundleName+"-application.zip"))

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

	dockerContextDir := "dockerContext"
	cleanContextFolder(dockerContextDir) // if was already present

	// copy application file in Docker build context folder:
	appName := filepath.Base(appPath)
	if err := cp.Copy(appPath, filepath.Join(dockerContextDir, appName)); err != nil {
		return err
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

	defer cleanContextFolder(dockerContextDir)

	fullDockerImageName := *tag
	if *tag == dockerImagePrefix {
		fullDockerImageName = *tag + edition
	}
	baseImageName := "bonitasoft.jfrog.io/docker-snapshot-local/bonita-community" // FIXME
	baseImageVersion := "8.0-SNAPSHOT"                                            // FIXME
	if *dockerSubscription {
		baseImageName = "docker.io/bonitasoft/bonita-subscription" // FIXME
		baseImageVersion = "latest"                                // FIXME
	}
	opts := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{fullDockerImageName},
		Remove:     true,
		BuildArgs: map[string]*string{
			"BONITA_IMAGE_NAME":       &baseImageName,
			"BONITA_IMAGE_VERSION":    &baseImageVersion,
			"CUSTOM_APPLICATION_FILE": &appName},
	}
	if *verbose {
		fmt.Println("Building image: " + fullDockerImageName)
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
	fmt.Printf("\nSuccessfully created Docker image '%s'\n\n", fullDockerImageName)
	return nil
}

func cleanContextFolder(dockerContextDir string) {
	if Exists(dockerContextDir) {
		if *verbose {
			fmt.Println("Cleaning temporary docker context folder")
		}
		if err := os.RemoveAll(dockerContextDir); err != nil {
			panic(err)
		}
	}
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
			if *verbose {
				fmt.Println("Creating zip dir", path)
			}
			_, err := w.Create(path + "/")
			if err != nil {
				return err
			}
			// then add files inside it
			if err := addFilesToZip(w, fullfilepath, baseInZip+"/"+file.Name()); err != nil {
				return err
			}
		} else if file.Mode().IsRegular() {
			if *verbose {
				fmt.Println("Adding zip file", filepath.Join(baseInZip, file.Name()))
			}
			dat, err := ioutil.ReadFile(fullfilepath)
			if err != nil {
				return err
			}
			fh := &zip.FileHeader{Name: baseInZip + "/" + file.Name()}
			fh.SetMode(file.Mode())
			f, err := w.CreateHeader(fh)
			// f, err := w.Create(filepath.Join(baseInZip, file.Name()))
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
