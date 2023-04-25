package packager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	zip "github.com/bonitasoft/bonita-application-packager/zip"
	cp "github.com/otiai10/copy"
)

func init() {
	tomcatPackageCmd.Flags().StringVarP(&tomcatBundleFile, "bonita-tomcat-bundle", "b", "",
		`(Optional) Specify path to the Bonita tomcat bundle file (Bonita*.zip) used to build.
If not passed, looking for a Bonita tomcat bundle in current folder`)
}

var (
	tomcatBundleFile string

	tomcatPackageCmd = &cobra.Command{
		Use: "tomcat [PATH_TO_YOUR_APPLICATION]",
		// ValidArgs: []string{"tomcat", "docker"},
		Short: "Package your Custom Application within a Bonita Tomcat 😺 Bundle",
		Long: `Package your Custom Application within a Bonita Tomcat 😺 Bundle.
The resulting package is self-contained and deploys itself entirely at Tomcat startup without further manual operations`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			applicationPath = args[0]
			buildTomcatBundle()
		},
	}
)

func buildTomcatBundle() {
	bundleAbsolutePath := validateTomcatBundle()
	if bundleAbsolutePath == "" {
		return
	}
	if exists("output") {
		if Verbose {
			fmt.Println("Cleaning 'output/' folder")
		}
		if err := os.RemoveAll("output"); err != nil {
			fmt.Println("Failed to clean 'output/' folder")
			panic(err)
		}
	}

	fmt.Println("Generating your Custom Application Bonita Tomcat 😺 Bundle...")

	bundleNameAndPath := filepath.Base(bundleAbsolutePath)                      // just the name of the zip file without the path
	bundleName := bundleNameAndPath[0:strings.Index(bundleNameAndPath, ".zip")] // just the name without '.zip'
	fmt.Printf("Unpacking Bonita Tomcat bundle %s.zip\n", bundleName)
	bundleRootDirInsideZip := zip.UnzipFile(bundleAbsolutePath, "output")
	if bundleRootDirInsideZip == "" {
		panic("No root folder found inside file " + bundleAbsolutePath)
	}
	fmt.Println("Unpacking Bonita WAR file")
	zip.UnzipFile(filepath.Join("output", bundleRootDirInsideZip, "server", "webapps", "bonita.war"), filepath.Join("output", bundleRootDirInsideZip, "server", "webapps", "bonita"))
	if Verbose {
		fmt.Println("Removing unpacked Bonita WAR file")
	}
	if err := os.Remove(filepath.Join("output", bundleRootDirInsideZip, "server", "webapps", "bonita.war")); err != nil {
		panic(err)
	}
	fmt.Println("Copying your custom application inside Bonita")
	copyResourceToCustomAppFolder(bundleRootDirInsideZip, applicationPath)
	if configurationFile != "" {
		fmt.Println("Copying your Bonita configuration file inside Bonita")
		copyResourceToCustomAppFolder(bundleRootDirInsideZip, configurationFile)
	}
	fmt.Println("Re-packing Bonita bundle containing your application")
	err := zip.ZipDirectory(filepath.Join("output", bundleName+"-application.zip"), filepath.Join("output", bundleRootDirInsideZip), bundleRootDirInsideZip)
	if err != nil {
		panic(err)
	}
	tempfolderToZip := filepath.Join("output", bundleRootDirInsideZip)
	if exists(tempfolderToZip) {
		if Verbose {
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
	if tomcatBundleFile != "" {
		if !exists(tomcatBundleFile) {
			fmt.Println("Bonita Tomcat bundle file passed as parameter does not exist: " + tomcatBundleFile)
			return ""
		} else if !strings.HasSuffix(tomcatBundleFile, ".zip") {
			fmt.Println("Bonita Tomcat bundle file passed as parameter is not a proper Bonita Tomcat bundle ZIP file: " + tomcatBundleFile)
			return ""
		} else {
			if Verbose {
				fmt.Println("Using Bonita Tomcat bundle file passed as parameter", tomcatBundleFile)
			}
			return tomcatBundleFile
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
			fmt.Println("or use parameter --bonita-tomcat-bundle <PATH_TO_TOMCAT_BUNDLE> if stored somewhere else.")
			fmt.Println("Then re-run this program")
			return ""
		}
		if Verbose {
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
