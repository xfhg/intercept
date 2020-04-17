package cmd

import (
	"fmt"
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
