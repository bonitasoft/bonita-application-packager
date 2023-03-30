package packager

import (
	"context"
	_ "embed"
	"fmt"
	"io"
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
	"github.com/spf13/cobra"
	t "golang.org/x/term"
)

const (
	defaultImageTag  = "my-bonita-application:latest"
	defaultBaseImage = "bonita:latest"
)

func init() {
	dockerPackageCmd.Flags().StringVarP(&tag, "tag", "t", defaultImageTag, "Docker image tag to use when building")
	dockerPackageCmd.Flags().StringVarP(&baseImage, "bonita-base-image", "i", defaultBaseImage, "Specify Bonita base docker image")
	dockerPackageCmd.Flags().StringVarP(&registryUsername, "registry-username", "u", "", "Specify username to authenticate against Bonita base docker image Registry")
	dockerPackageCmd.Flags().StringVarP(&registryPassword, "registry-password", "p", "", `Specify password to authenticate against Bonita base docker image Registry
If --registry-username is provided and not --registry-password, password will be prompted interactively and never issued to the console`)
}

var (
	tag              string
	baseImage        string
	registryUsername string
	registryPassword string

	dockerPackageCmd = &cobra.Command{
		Use:   "docker",
		Short: "Package your Custom Application inside a Bonita Docker 🐳 Image",
		Long: fmt.Sprintf(
			`Package your Custom Application within a Bonita Docker 🐳 image.
The resulting package is self-contained and deploys itself entirely at Docker container startup without further manual operations,

use --tag to specify the name of your resulting built image
use --bonita-base-image to specify a Bonita docker base image different from the default, which is '%s' (community version)
    Access your account on Bonita Artifact Repository to see the list of available docker base images
    (https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access) 
use --registry-username and --registry-password if you need to authenticate against the docker image registry to pull Bonita docker base image`, defaultBaseImage),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationPath = args[0]
			buildDockerImage()
		},
	}

	// Go directive to include Dockerfile in the binary:
	//go:embed Dockerfile
	dockFile []byte
)

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
	fmt.Println("Generating your Custom Application Bonita Docker 🐳 Image...")
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
	appName := filepath.Base(applicationPath)
	if err := cp.Copy(applicationPath, filepath.Join(dockerResourcesDir, appName)); err != nil {
		return err
	}

	bconfName := ""
	if configurationFile != "" {
		// copy bconf file in Docker build context resources folder:
		bconfName = filepath.Base(configurationFile)
		if err := cp.Copy(configurationFile, filepath.Join(dockerResourcesDir, bconfName)); err != nil {
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
	fmt.Printf("\nSuccessfully created Docker image '%s'\n\n", tag)
	fmt.Printf("To use it, run appropriate command:\n")
	fmt.Printf("- Community release    : docker run --name my-bonita-app -d -p 8080:8080 %s\n", tag)
	fmt.Printf("- Subscription release : docker run --name my-bonita-app -h <hostname> -v <license-folder>:/opt/bonita_lic/ -d -p 8080:8080 %s\n", tag)
	fmt.Printf("Read https://documentation.bonitasoft.com/bonita/latest/runtime/bonita-docker-installation for complete options on how to run a Bonita-based Docker container.\n")
	return nil
}

func buildCustomDockerImage(ctx context.Context, dockerClient *client.Client, dockerContext io.ReadCloser) error {
	dockerfile := "Dockerfile"
	opts := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{tag},
		Remove:     true,
		BuildArgs: map[string]*string{
			"BONITA_BASE_IMAGE": &baseImage},
	}
	if Verbose {
		fmt.Println("Building new image:", tag)
	}

	if registryUsername != "" && registryPassword == "" {
		fmt.Printf("Enter your password to access the Docker Registry corresponding to '%v':", registryUsername)
		p, err := t.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("Error reading your password")
			return err
		}
		registryPassword = string(p)
		fmt.Println() // to make the next print on a fresh new line
	}

	// configure registry authentication
	if registryUsername != "" && registryPassword != "" {
		registryName, _, found := strings.Cut(baseImage, "/")
		if !found {
			// if no registry found, set default docker registry
			registryName = "docker.io"
		}
		if Verbose {
			fmt.Println("Authenticating to registry:", registryName)
		}
		opts.AuthConfigs = map[string]types.AuthConfig{
			registryName: {
				Username: registryUsername,
				Password: registryPassword,
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
