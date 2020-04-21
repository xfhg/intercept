package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/kardianos/osext"
	homedir "github.com/mitchellh/go-homedir"
)

// FileExists check if file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// CoreExists return path of core binaries on this platform
func CoreExists() string {

	rgbin := ""

	executablePath := GetExecutablePath()

	switch runtime.GOOS {
	case "windows":
		rgbin = "/rg/rg.exe"
	case "darwin":
		rgbin = "/rg/rgm"
	case "linux":
		rgbin = "/rg/rgl"
	default:
		colorRedBold.Println("| OS not supported")
		PrintClose()
		os.Exit(1)
	}

	fullcorePath := executablePath + rgbin

	if !FileExists(fullcorePath) {
		colorRedBold.Println("| RG not found")
		colorRedBold.Println("| Run the command - intercept system - ")
		PrintClose()
		os.Exit(1)
	}
	return fullcorePath

}

// PrintStart prints the command banner
func PrintStart() {
	fmt.Println("┌")
	fmt.Println("| INTERCEPT")
	fmt.Println("|")
}

// ContainsInt checks for ints
func ContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// PrintClose prints the command ending
func PrintClose() {

	fmt.Println("|")
	fmt.Println("|")
	fmt.Println("| INTERCEPT")
	fmt.Println("|", buildVersion)
	fmt.Println("└")
	fmt.Println("")

}

// ReaderFromURL grabs de config from URL
func ReaderFromURL(path string) (io.ReadCloser, error) {
	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return res.Body, nil
	}

	fmt.Println("|")
	fmt.Println("| Could not download config file")
	fmt.Println("└")
	os.Exit(0)

	return nil, err

}

// GetHomeDir returns home directory
func GetHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		PrintClose()
		os.Exit(1)
	}
	return home
}

// GetExecutablePath returns where the main executable is running from
func GetExecutablePath() string {

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	return folderPath

}

// GetWd returns working directory
func GetWd() string {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir

}
