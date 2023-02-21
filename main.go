package main

import (
	"archive/zip"
	"bufio"
	"context"
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
	cp "github.com/otiai10/copy"
)

const (
	_sep              = os.PathSeparator
	dockerImagePrefix = "bonitasoft/bonita-package-docker-"
)

var (
	// Flags:
	tomcatFlag = flag.Bool("tomcat", false, "Choose to build a Tomcat bundle containing your application")
	dockerFlag = flag.Bool("docker", false, `Choose to build a docker image containing your application,
	use -tag to specify the name of your built image
	By default, it builds a 'Community' Docker image
	use -subscription to build a 'Subscription" Docker image (you must have the rights to download Bonita Subscription Docker base image)`)
	dockerSubscription = flag.Bool("subscription", false, "Choose to build a Subscription-based docker image (default build a Community image)")
	tag                = flag.String("tag", dockerImagePrefix, "Docker image tag to use when building")
	verbose            = flag.Bool("verbose", false, "Enable verbose (debug) mode")
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintln(w, "This tool allows to build a Bonita Tomcat bundle or a Bonita Docker image containing your custom application")
		fmt.Fprintln(w, "one of '-tomcat' or '-docker' is mandatory")
		flag.PrintDefaults()
	}

	// Create src/my-application folder if not exist:
	myAppFolder := filepath.Join("src", "my-application")
	if !Exists(myAppFolder) {
		if err := os.MkdirAll(myAppFolder, os.ModePerm); err != nil {
			panic(err)
		}
	}

	if !*dockerFlag && !*tomcatFlag {
		fmt.Printf("Please specify '-tomcat' if you want to build a Bonita Tomcat Bundle or '-docker' if you want to build a Bonita Docker image\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Verbose mode          :", *verbose)
	fmt.Println("Build Tomcat bundle   :", *tomcatFlag)
	fmt.Println("Build Docker image    :", *dockerFlag)
	var dockerEdition = "community"
	if *dockerFlag {
		fmt.Println("Docker image tag name :", *tag)
		if *dockerSubscription {
			dockerEdition = "subscription"
		}
		fmt.Println("Docker image edition  :", dockerEdition)
	}

	// Check that application ZIP is present
	applicationMatches, err := filepath.Glob("src" + string(_sep) + "my-application" + string(_sep) + "*.zip")
	if err != nil {
		panic(err)
	}

	if *tomcatFlag {
		buildTomcatBundle(applicationMatches)
	}

	if *dockerFlag {
		fmt.Println("Building Docker image (Community & Subscription editions)")
		buildDockerImages(dockerEdition)
	}
}

func buildTomcatBundle(applicationMatches []string) {
	// Try to find a Bonita zip file inside:
	bundleMatches, err := filepath.Glob("src" + string(_sep) + "Bonita*.zip")
	if err != nil {
		panic(err)
	}
	if len(bundleMatches) == 0 || len(applicationMatches) == 0 {
		fmt.Println("Please place:")
		fmt.Println(" * in src/ folder, the ZIP file of Bonita Tomcat Bundle (Eg. BonitaCommunity-2023.1-u0.zip, BonitaSubscription-2023.1-u2.zip)")
		fmt.Println(" * in src/my-application/ folder the ZIP file containing your entire application")
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
	bundleName := bundleNameAndPath[4:strings.Index(bundleNameAndPath, ".zip")] // until end of string
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
	err = cp.Copy(filepath.Join("src", "my-application"), filepath.Join("output", bundleName, "server", "webapps", "bonita", "WEB-INF", "classes", "my-application"))
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

func buildDockerImages(dockerEdition string) {
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

	tar, err := archive.TarWithOptions("src/", &archive.TarOptions{})
	if err != nil {
		return err
	}

	fullDockerImageName := *tag + edition
	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile." + edition,
		Tags:       []string{fullDockerImageName},
		Remove:     true,
		// BuildArgs: ,
	}
	if *verbose {
		fmt.Println("Dockerfile used: " + opts.Dockerfile)
		fmt.Println("Building image: " + fullDockerImageName)
	}

	res, err := dockerClient.ImageBuild(ctx, tar, opts)
	if err != nil {
		fmt.Println("Error building image")
		return err
	}

	defer res.Body.Close()

	err = print(res.Body)
	if err != nil {
		return err
	}
	fmt.Printf("\nSuccessfully created Docker image '%s'\n\n", fullDockerImageName)

	return nil
}

func print(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		if *verbose { // print docker build output if verbose mode ON:
			fmt.Println(scanner.Text())
		}
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
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
		//fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			//fmt.Println("creating directory...")
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
			if err := addFilesToZip(w, fullfilepath, filepath.Join(baseInZip, file.Name())); err != nil {
				return err
			}
		} else if file.Mode().IsRegular() {
			dat, err := ioutil.ReadFile(fullfilepath)
			// fileRead, err := os.Open(fullfilepath) // dat, err := ioutil.ReadFile(fullfilepath)
			if err != nil {
				return err
			}
			//w.OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, file.Mode())
			fh := &zip.FileHeader{Name: filepath.Join(baseInZip, file.Name())}
			fh.SetMode(file.Mode())
			f, err := w.CreateHeader(fh)
			// f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			// _, err = io.Copy(f, fileRead) //f.Write(dat)
			if err != nil {
				return err
			}
		} else {
			// we ignore non-regular files because they are scary
		}
	}
	return nil
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}
