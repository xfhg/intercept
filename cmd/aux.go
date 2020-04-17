package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// FileExists check if file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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
	fmt.Println("| INTERCEPT")
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
