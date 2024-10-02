package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"bytes"

	"cuelang.org/go/cue"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

func patchJSONOutputFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening JSON file: %v", err)
	}
	defer file.Close()

	var validObjects []map[string]interface{}

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to parse each line as a separate JSON object
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			validObjects = append(validObjects, obj)
		} else {
			// Log the error and skip this line
			fmt.Printf("Skipping invalid JSON line: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Marshal the array of objects into a properly formatted JSON string
	validJSON, err := json.MarshalIndent(validObjects, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Write the patched JSON back to the file
	err = os.WriteFile(filePath, validJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing patched JSON to file: %v", err)
	}

	return nil
}

func NormalizeFilename(input string) string {
	// Convert to lowercase and replace spaces/underscores with hyphens
	processed := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return '-'
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			return unicode.ToLower(r)
		}
		return -1 // remove non-alphanumeric characters
	}, input)

	return processed
}

func NormalizePolicyName(input string) string {
	// Convert to lowercase and replace spaces/underscores with hyphens
	processed := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return '-'
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			return unicode.ToUpper(r)
		}
		return -1 // remove non-alphanumeric characters
	}, input)

	return processed
}

// calculateSHA256 calculates the SHA256 hash of a file
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

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func findMissingFields(schema, data cue.Value) []string {
	var missingFields []string
	findMissingFieldsRecursive(schema, data, "", &missingFields)
	return missingFields
}

func findMissingFieldsRecursive(schema, data cue.Value, path string, missingFields *[]string) {
	iter, _ := schema.Fields()
	for iter.Next() {
		label := iter.Label()
		schemaField := iter.Value()
		dataField := data.LookupPath(cue.ParsePath(label))

		currentPath := joinPath(path, label)

		if !dataField.Exists() {
			*missingFields = append(*missingFields, currentPath)
		} else if schemaField.Kind() == cue.StructKind && dataField.Kind() == cue.StructKind {
			findMissingFieldsRecursive(schemaField, dataField, currentPath, missingFields)
		}
	}
}

func findExtraFields(schema, data cue.Value) []string {
	var extraFields []string
	findExtraFieldsRecursive(schema, data, "", &extraFields)
	return extraFields
}

func findExtraFieldsRecursive(schema, data cue.Value, path string, extraFields *[]string) {
	iter, _ := data.Fields()
	for iter.Next() {
		label := iter.Label()
		dataField := iter.Value()
		schemaField := schema.LookupPath(cue.ParsePath(label))

		currentPath := joinPath(path, label)

		if !schemaField.Exists() {
			*extraFields = append(*extraFields, currentPath)
		} else if dataField.Kind() == cue.StructKind && schemaField.Kind() == cue.StructKind {
			findExtraFieldsRecursive(schemaField, dataField, currentPath, extraFields)
		}
	}
}

func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return fmt.Sprintf("%s.%s", base, field)
}

func validateSchema(schema, data cue.Value) ([]string, []string) {
	missingFields := findMissingFields(schema, data)
	extraFields := findExtraFields(schema, data)

	// Handle top-level mismatch
	if !schema.Subsumes(data) {
		schemaKeys := getKeys(schema)
		dataKeys := getKeys(data)

		for _, key := range schemaKeys {
			if !contains(dataKeys, key) && !contains(missingFields, key) {
				missingFields = append(missingFields, key)
			}
		}

		for _, key := range dataKeys {
			if !contains(schemaKeys, key) && !contains(extraFields, key) {
				extraFields = append(extraFields, key)
			}
		}
	}

	return missingFields, extraFields
}

func getKeys(v cue.Value) []string {
	var keys []string
	iter, _ := v.Fields()
	for iter.Next() {
		keys = append(keys, iter.Label())
	}
	return keys
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item || strings.HasPrefix(item, s+".") {
			return true
		}
	}
	return false
}

func convertToJSON(content []byte, contentType string) ([]byte, error) {
	switch contentType {
	case "json":
		return content, nil
	case "yaml":
		var yamlObj interface{}
		if err := yaml.Unmarshal([]byte(content), &yamlObj); err != nil {
			return nil, err
		}
		return json.Marshal(yamlObj)
	case "ini":
		cfg, err := ini.Load(content)
		if err != nil {
			return nil, err
		}
		return []byte(iniToJSONLike(cfg)), nil
	case "toml":
		var tomlObj map[string]interface{}
		if err := toml.Unmarshal([]byte(content), &tomlObj); err != nil {
			return nil, err
		}
		return json.Marshal(tomlObj)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func convertFromJSON(content []byte, coutputType string) ([]byte, error) {
	var jsonObj interface{}
	if err := json.Unmarshal(content, &jsonObj); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	switch coutputType {
	case "json":
		return content, nil
	case "yaml":
		return yaml.Marshal(jsonObj)
	case "ini":
		cfg := ini.Empty()
		err := convertJSONToINI(jsonObj, cfg, "")
		if err != nil {
			return nil, fmt.Errorf("error converting JSON to INI: %w", err)
		}
		var buf bytes.Buffer
		_, err = cfg.WriteTo(&buf)
		if err != nil {
			return nil, fmt.Errorf("error writing INI to buffer: %w", err)
		}
		return buf.Bytes(), nil
	case "toml":
		return convertJSONToTOML(jsonObj)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", coutputType)
	}
}

func convertJSONToINI(jsonObj interface{}, cfg *ini.File, section string) error {
	switch v := jsonObj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if section == "" {
				if err := convertJSONToINI(value, cfg, key); err != nil {
					return err
				}
			} else {
				sec, err := cfg.NewSection(section)
				if err != nil {
					return err
				}
				if _, err := sec.NewKey(key, fmt.Sprintf("%v", value)); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unexpected type in JSON structure")
	}
	return nil
}

func convertJSONToTOML(jsonObj interface{}) ([]byte, error) {
	// Convert the data to preserve integers
	convertedData := convertData(jsonObj)

	// Marshal to TOML
	tomlData, err := toml.Marshal(convertedData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to TOML: %w", err)
	}

	return tomlData, nil
}

func convertData(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[k] = convertData(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = convertData(val)
		}
		return v
	case float64:
		// Check if the float64 is actually an integer
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	case json.Number:
		if i, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(string(v), 64); err == nil {
			return f
		}
		return v.String()
	default:
		return v
	}
}

func createOutputDirectories(isObserve bool) error {
	dirs := []string{"_patched", "_debug", "_sarif"}
	if isObserve {
		dirs = []string{"_patched", "_debug", "_sarif", "_status"}
	}
	for _, dir := range dirs {
		if outputDir != "" {
			dir = filepath.Join(outputDir, dir)
			// log.Debug().Msgf("Creating directory: %s", dir)

		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error().Err(err).Msgf("failed to create directory %s", dir)
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func cleanupOutputDirectories() error {
	dirsToClean := []string{"_sarif", "_debug", "_patched"}
	if outputDir != "" {
		for i, dir := range dirsToClean {
			dirsToClean[i] = filepath.Join(outputDir, dir)
			// log.Debug().Msgf("Cleaning up directories: %v", dirsToClean)
		}
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(dirsToClean))

	for _, dir := range dirsToClean {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			err := os.RemoveAll(d)
			if err != nil {
				errChan <- fmt.Errorf("failed to clean directory %s: %w", d, err)
				return
			}

			// log.Debug().Msgf("Cleaned up directory: %s ", d)
		}(dir)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		// Path exists
		return true
	}
	if os.IsNotExist(err) {
		// Path does not exist
		return false
	}
	// Error is due to something else, such as a permission issue
	return false
}

func PathInfo(path string) (exists bool, isDir bool, err error) {
	info, err := os.Stat(path)
	if err == nil {
		// Path exists; check if it is a directory
		return true, info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		// Path does not exist
		return false, false, nil
	}
	// Error due to something else, such as a permission issue
	return false, false, err
}

func GetDirectory(path string) string {
	return filepath.Dir(path) + "/"
}

// isURL checks if the input string is a valid URL
func isURL(input string) bool {
	// Check for common URL schemes
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		_, err := url.ParseRequestURI(input)
		return err == nil
	}
	return false
}
