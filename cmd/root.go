package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pkg "github.com/bonitasoft/bonita-application-packager/cmd/packager"
)

var (
	// Version is overridden via build flag: -ldflags="-X 'github.com/bonitasoft/bonita-application-packager/cmd.Version=0.0.8'"
	// convention is <module_name>/<package>/<sub-package>.<VarNameToOverride>
	Version = "development"

	RootCmd = &cobra.Command{
		Use:   "bonita <command> [flags]",
		Short: "Bonita CLI",
		Long: `This tool allows to build a Bonita Tomcat bundle or a Bonita Docker image containing your custom application.

Complete usage available at https://github.com/bonitasoft/bonita-application-packager`,

		Example: `$ bonita package tomcat /path/to/my-application.zip
$ bonita package docker /path/to/my-application.zip`,
		// This sets the version support by Cobra (adds --version flag):
		Version: Version,
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// example of a flag (valid for all (sub-)commands):
	// RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More verbose information output")

	RootCmd.AddCommand(pkg.PackageCmd)
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetVersionTemplate(fmt.Sprintf("Bonita CLI %s\n", Version))
}
