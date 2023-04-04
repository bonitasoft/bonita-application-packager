package packager

import (
	"fmt"
	"os"
)

func printFinalNote(additionalNote string) {
	fmt.Println("\nNOTE: if your custom application is using pages from Bonita Admin or User applications, " +
		additionalNote + " in order to install those pages, else, your application will fail at install.")
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func printMsgIfVerbose[T any](message string, o T) T {
	if Verbose {
		fmt.Println(message, o)
	}
	return o
}
