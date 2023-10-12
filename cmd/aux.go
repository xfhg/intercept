package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kardianos/osext"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
)

func cleanupFiles() {

	_ = os.Remove("intercept.output.json")
	_ = os.Remove("intercept.sarif.json")
	_ = os.Remove("intercept.scannedSHA256.json")

}

func LazyMatch(allJSONData map[string]interface{}, jsonPiece map[string]interface{}) bool {
	for key, pieceValue := range jsonPiece {
		if allValue, ok := allJSONData[key]; ok {
			switch pieceTyped := pieceValue.(type) {
			case map[string]interface{}:
				allValueMap, ok := allValue.(map[string]interface{})
				if !ok || !LazyMatch(allValueMap, pieceTyped) {
					return false
				}
			case []interface{}:
				allValueArray, ok := allValue.([]interface{})
				if !ok {
					return false
				}
				found := false
				for _, entry := range allValueArray {
					entryMap, ok := entry.(map[string]interface{})
					if !ok {
						continue
					}
					for _, pieceEntry := range pieceTyped {
						pieceEntryMap, ok := pieceEntry.(map[string]interface{})
						if !ok {
							continue
						}
						if LazyMatch(entryMap, pieceEntryMap) {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					return false
				}
			default:
				if allValue != pieceValue {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func isSubsetOrEqual(a, b map[string]interface{}) bool {
	for k, v := range a {
		if bv, exists := b[k]; exists {
			if reflect.DeepEqual(v, bv) {
				continue
			}
			switch vv := v.(type) {
			case map[string]interface{}:
				bvv, ok := bv.(map[string]interface{})
				if !ok || !isSubsetOrEqual(vv, bvv) {
					return false
				}
			case []interface{}:
				bvv, ok := bv.([]interface{})
				if !ok || !reflect.DeepEqual(vv, bvv) {
					return false
				}
			default:
				if v != bv {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func detectFormat(data string) string {
	// Try to unmarshal as JSON
	var js map[string]interface{}
	if json.Unmarshal([]byte(data), &js) == nil {
		return "application/json"
	}

	// Try to unmarshal as XML
	var xm map[string]interface{}
	if xml.Unmarshal([]byte(data), &xm) == nil {
		return "application/xml"
	}

	return "unknown"
}

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

func getJSONRootKeys(jsonObj map[string]interface{}) []string {
	var keys []string
	for k := range jsonObj {
		keys = append(keys, k)
	}
	return keys
}

func isTOMLKeyAbsent(tree *toml.Tree, key string) bool {
	return tree.Get(key) == nil
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

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashSum := hash.Sum(nil)
	hashString := hex.EncodeToString(hashSum)

	return hashString, nil
}

// func unmarshalYAML(data []byte) (map[string]interface{}, error) {
// 	var result map[string]interface{}
// 	err := yaml.Unmarshal(data, &result)
// 	return result, err
// }

func PathSHA256(root string) ([]ScannedFile, error) {
	var files []ScannedFile
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			hash := sha256.New()
			if _, err := io.Copy(hash, file); err != nil {
				return err
			}

			files = append(files, ScannedFile{
				Path:   path,
				Sha256: fmt.Sprintf("%x", hash.Sum(nil)),
			})
		}
		return nil
	})
	return files, err
}

// func DownloadJSONFile(src, dst string) error {
// 	// Send a HEAD request to get the Content-Type without downloading the body
// 	resp, err := http.Head(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	// Check if the Content-Type is JSON
// 	// if resp.Header.Get("Content-Type") != "application/json" {
// 	// 	return errors.New("source is not a JSON file")
// 	// }

// 	// Configure the getter client
// 	client := &getter.Client{
// 		Src:  src,
// 		Dst:  dst,
// 		Mode: getter.ClientModeFile,
// 	}

// 	// Configure the yacspin spinner
// 	cfg := yacspin.Config{
// 		Frequency:       100 * time.Millisecond,
// 		CharSet:         yacspin.CharSets[3],
// 		Suffix:          " Downloading",
// 		SuffixAutoColon: true,
// 		StopCharacter:   "│ ✓",
// 		StopColors:      []string{"fgGreen"},
// 	}

// 	spinner, err := yacspin.New(cfg)
// 	if err != nil {
// 		return err
// 	}

// 	// Start the spinner
// 	err = spinner.Start()
// 	if err != nil {
// 		return err
// 	}

// 	// Download the file
// 	err = client.Get()
// 	if err != nil {
// 		spinner.StopFail()
// 		return err
// 	}

// 	// Stop the spinner
// 	spinner.Stop()

// 	// Read the downloaded file
// 	data, err := ioutil.ReadFile(dst)
// 	if err != nil {
// 		return err
// 	}

// 	// Compute the SHA-256 hash of the file
// 	hash := sha256hash(data)

// 	fmt.Println("│")
// 	fmt.Println("├ Hash: ", hash)
// 	fmt.Println("│")

// 	return nil
// }
