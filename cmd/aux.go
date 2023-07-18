package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kardianos/osext"
	homedir "github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// FileExists check if file exists
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// // CoreExists return path of core binaries on this platform
// func CoreExists() string {

// 	rgbin := ""

// 	executablePath := GetExecutablePath()

// 	switch runtime.GOOS {
// 	case "windows":
// 		rgbin = filepath.Join("rg", "rg.exe")
// 	case "darwin":
// 		rgbin = filepath.Join("rg", "rgm")
// 	case "linux":
// 		rgbin = filepath.Join("rg", "rgl")
// 	default:
// 		colorRedBold.Println("│ OS not supported")
// 		PrintClose()
// 		os.Exit(1)
// 	}

// 	fullcorePath := filepath.Join(executablePath, rgbin)

// 	if !FileExists(fullcorePath) {
// 		colorRedBold.Println("│ RG not found")
// 		colorRedBold.Println("│ Run the command - intercept system - ")
// 		PrintClose()
// 		os.Exit(1)
// 	}
// 	return fullcorePath

// }

// PrintStart prints the command banner
func PrintStart() {
	fmt.Println("┌")
	fmt.Println("│ INTERCEPT")
	fmt.Println("│")
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

func FindMatchingString(s1 string, s2 string, delim string) bool {

	s1 = strings.ToUpper(s1)
	s2 = strings.ToUpper(s2)

	a1 := strings.Split(s1, delim)
	a2 := strings.Split(s2, delim)
	for _, str1 := range a1 {
		for _, str2 := range a2 {
			if str1 == str2 {
				return true
			}
		}
	}
	return false
}

// PrintClose prints the command ending
func PrintClose() {

	fmt.Println("│")
	fmt.Println("│")
	colorBlueBold.Println("├ INTERCEPT")
	fmt.Println("│ https://intercept.cc")
	fmt.Println("│")
	if buildVersion != "" {
		fmt.Println("├", buildVersion)
	}
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

	fmt.Println("│")
	fmt.Println("│ Could not download config file")
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

// WriteLinesOnFile does that
func WriteLinesOnFile(lines []string, filepath string) error {

	var err error
	var f *os.File

	if !FileExists(filepath) {

		f, err = os.Create(filepath)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return err
		}

		for _, v := range lines {
			fmt.Fprintln(f, v)
			if err != nil {
				log.Fatal(err)
				return err
			}
		}
		err = f.Close()
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return err
}

// LogError does that
func LogError(err error) {
	colorRedBold.Println("│")
	colorRedBold.Println("│ Error: ")
	colorRedBold.Println("│ ", err)
	colorRedBold.Println("│")
	PrintClose()
	log.Fatal(err)
}

func sha256hash(data []byte) string {

	HexDigest := sha256.Sum256(data)
	return hex.EncodeToString(HexDigest[:])
}

func unmarshalYAML(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal(data, &result)
	return result, err
}
