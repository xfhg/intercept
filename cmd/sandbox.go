package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var targetFile string
var experimentalInput string

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "INTERCEPT / SANDBOX - experimental features",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("│")
		fmt.Println("├ Sandbox")
		fmt.Println("│")

		// dir := experimentalInput

		// // Regex pattern to match files
		// regexPattern := "^development-\\d+\\.yml$"

		// schema := `ingress: {enabled: true}
		// service: {port: 88}
		// 	...
		// 	`

		// // Compile regex pattern
		// regex, err := regexp.Compile(regexPattern)
		// if err != nil {
		// 	fmt.Printf("Error compiling regex: %s\n", err)
		// 	return
		// }

		// // directory check

		// fileInfo, err := os.Stat(dir)
		// if err != nil {
		// 	log.Fatalf("Error accessing path: %s", err)
		// }

		// // Check if the path is a directory
		// if fileInfo.IsDir() {
		// 	fmt.Printf("%s is a directory\n", dir)
		// } else {
		// 	fmt.Printf("%s is not a directory\n", dir)
		// }

		// // Function to match files
		// err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// 	if err != nil {
		// 		return err
		// 	}
		// 	// Ignore directories and match files against regex pattern
		// 	if !info.IsDir() && regex.MatchString(info.Name()) {
		// 		fmt.Println(path) // Print matched file path
		// 		filehash, _ := calculateSHA256(path)
		// 		fmt.Println(filehash)

		// 		// Read the file in path to a string ymlContent

		// 		ymlContentBytes, err := os.ReadFile(path)
		// 		if err != nil {
		// 			fmt.Printf("Error reading file: %s\n", err)
		// 			return nil
		// 		}

		// 		var yamlObj interface{}
		// 		err = yaml.Unmarshal(ymlContentBytes, &yamlObj)
		// 		if err != nil {
		// 			fmt.Printf("error unmarshaling YAML data: %v", err)
		// 			return nil
		// 		}

		// 		ymlContent := string(ymlContentBytes)

		// 		xvalid, xerrMsg := validateYAMLAndCUEContent(ymlContent, schema)
		// 		if xvalid {
		// 			fmt.Printf("The YAML file is valid.\n")
		// 		} else {
		// 			fmt.Printf("The YAML file is not valid: %s\n", xerrMsg)
		// 		}

		// 	}
		// 	return nil
		// })
		// if err != nil {
		// 	fmt.Printf("Error walking directory: %s\n", err)
		// }

		PrintClose()

	},
}

func init() {

	sandboxCmd.PersistentFlags().BoolP("activate", "l", false, "Activate Sandbox Features")

	sandboxCmd.PersistentFlags().StringVarP(&targetFile, "dl", "d", "", "Download target file")
	sandboxCmd.PersistentFlags().StringVarP(&experimentalInput, "xi", "j", "", "Experimental input")

	rootCmd.AddCommand(sandboxCmd)

}

// func validateYAMLAgainstCUE(yamlFile string, cueFile string) (bool, string) {
// 	yamlData, err := os.ReadFile(yamlFile)
// 	if err != nil {
// 		return false, fmt.Sprintf("error reading YAML file: %v", err)
// 	}

// 	var yamlObj interface{}
// 	err = yaml.Unmarshal(yamlData, &yamlObj)
// 	if err != nil {
// 		return false, fmt.Sprintf("error unmarshaling YAML data: %v", err)
// 	}

// 	var r cue.Runtime
// 	binst := load.Instances([]string{cueFile}, &load.Config{})
// 	if len(binst) != 1 || binst[0].Err != nil {
// 		return false, fmt.Sprintf("error loading CUE file: %v", binst[0].Err)
// 	}

// 	cueInstance, err := r.Build(binst[0])
// 	if err != nil {
// 		return false, fmt.Sprintf("error building CUE instance: %v", err)
// 	}

// 	cueValue := cueInstance.Value()
// 	err = cueValue.Validate(cue.Concrete(true))
// 	if err != nil {
// 		return false, fmt.Sprintf("error validating CUE value: %v", err)
// 	}

// 	jsonData, err := json.Marshal(yamlObj)
// 	if err != nil {
// 		return false, fmt.Sprintf("error marshaling YAML object to JSON: %v", err)
// 	}

// 	yamlCueValue, err := r.Compile("", string(jsonData))
// 	if err != nil {
// 		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
// 	}

// 	err = cueValue.Unify(yamlCueValue.Value()).Validate(cue.Concrete(true))
// 	if err != nil {
// 		return false, fmt.Sprintf("error validating YAML data against CUE schema: %v", err)
// 	}

// 	return true, ""
// }

func validateYAMLAndCUEContent(yamlContent string, cueContent string) (bool, string) {
	var yamlObj interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &yamlObj)
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling YAML data: %v", err)
	}

	var r cue.Runtime
	binst, err := r.Compile("", cueContent)
	if err != nil {
		return false, fmt.Sprintf("error compiling CUE content: %v", err)
	}

	cueValue := binst.Value()
	err = cueValue.Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating CUE value: %v", err)
	}

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling YAML object to JSON: %v", err)
	}

	yamlCueValue, err := r.Compile("", string(jsonData))
	if err != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
	}

	err = cueValue.Unify(yamlCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating YAML data against CUE schema: %v", err)
	}

	return true, ""
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
