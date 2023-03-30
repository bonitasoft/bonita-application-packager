package packager

import (
	"github.com/spf13/cobra"
)

func init() {
	PackageCmd.AddCommand(tomcatPackageCmd)
	PackageCmd.AddCommand(dockerPackageCmd)

	PackageCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "More verbose information output")
	PackageCmd.PersistentFlags().StringVarP(&configurationFile, "configuration-file", "c", "",
		`(Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)`)
}

var (
	Verbose           bool
	configurationFile string
	applicationPath   string

	PackageCmd = &cobra.Command{
		Use:       "package",
		ValidArgs: []string{"tomcat", "docker"},
		Short:     "Package 📦 your Custom Application with Bonita",
		Long: `Package 📦 your Custom Application within a Bonita Tomcat Bundle or a Bonita Docker image.
The resulting package is self-contained and deploys itself entirely at startup without further manual operations`,
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),

		Example: `$ bonita package tomcat /path/to/my-application.zip -b /path/to/bonita-bundle.zip
$ bonita package docker /path/to/my-application.zip --tag my-company/bonita-app:1.0.0`,
	}
)
