package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type HostInfo struct {
	Hostname     string
	OS           string
	Architecture string
	IPs          []string
	MAC          string
}

var (
	testOutputDir bool
)

var testEmbeddedCmd = &cobra.Command{
	Use:   "sys",
	Short: "Test intercept embedded core binaries",
	Long:  `This command extracts and runs the embedded rg and goss binaries to verify they are working correctly.`,
	Run: func(cmd *cobra.Command, args []string) {

		if testOutputDir {
			log.Debug().Msg("Checking OS Permissions:")
			if outputDir != "" {
				absPath, err := filepath.Abs(outputDir)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to get absolute path")
				}
				var paths []string
				paths = append(paths, absPath)
				hasPermissions, err := Permissions(paths)
				if err != nil {
					log.Error().Err(err).Msg("Permission check failed")
				} else if hasPermissions {
					log.Debug().Msg("You have sufficient permissions to create directories and write files on the target paths.")
				} else {
					log.Error().Msg("You do not have sufficient permissions on the target paths.")
				}
			} else {
				log.Error().Msg("add your intended -o <output directory> for verification")
			}
		}

		log.Debug().Msg("Core binaries:")
		log.Debug().Msgf("rg path: %s", rgPath)
		log.Debug().Msgf("goss path: %s", gossPath)

		files, err := os.ReadDir(filepath.Dir(rgPath))
		if err != nil {
			log.Debug().Msgf("Error reading directory %s: %v", filepath.Dir(rgPath), err)
			return
		}
		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				log.Debug().Msgf("Error getting info for file %s: %v", file.Name(), err)
				continue
			}
			log.Debug().Msgf("File: %s, Permissions: %s", info.Name(), info.Mode().Perm())
		}

		files, err = os.ReadDir(filepath.Dir(gossPath))
		if err != nil {
			log.Debug().Msgf("Error reading directory %s: %v", filepath.Dir(rgPath), err)
			return
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				log.Debug().Msgf("Error getting info for file %s: %v", file.Name(), err)
				continue
			}
			log.Debug().Msgf("File: %s, Permissions: %s", info.Name(), info.Mode().Perm())
		}

		// Test rg
		log.Debug().Msg("Testing rg:")
		rgCmd := exec.Command(rgPath, "--version")
		rgCmd.Stdout = os.Stdout
		rgCmd.Stderr = os.Stderr
		if err := rgCmd.Run(); err != nil {
			log.Debug().Msgf("Error running rg: %v", err)
		}

		// Test goss
		log.Debug().Msg("Testing goss:")
		gossCmd := exec.Command(gossPath, "--version")
		gossCmd.Stdout = os.Stdout
		gossCmd.Stderr = os.Stderr
		if err := gossCmd.Run(); err != nil {
			log.Debug().Msgf("Error running goss: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testEmbeddedCmd)

	testEmbeddedCmd.Flags().BoolVar(&testOutputDir, "permissions", false, "Check OS permissions with --output-dir or -o")
}

func GetHostInfo() (*HostInfo, error) {
	hostInfo := &HostInfo{}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}
	hostInfo.Hostname = hostname

	// Get OS and architecture
	hostInfo.OS = runtime.GOOS
	hostInfo.Architecture = runtime.GOARCH

	// Get IPs and MAC addresses
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // ignore interfaces that are down
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get addresses for interface %v: %v", iface.Name, err)
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, fmt.Errorf("failed to parse IP address %v: %v", addr.String(), err)
			}

			if ip.IsLoopback() {
				continue // ignore loopback addresses
			}

			hostInfo.IPs = append(hostInfo.IPs, ip.String())
		}
		// main MAC
		if iface.Flags&net.FlagUp != 0 && iface.HardwareAddr.String() != "" {
			hostInfo.MAC = iface.HardwareAddr.String()
		}

	}

	return hostInfo, nil
}

// FingerprintHost generates a fingerprint for the host using its identifiable information
func FingerprintHost(hostInfo *HostInfo) (string, string, string, error) {
	data := strings.Join([]string{
		hostInfo.MAC,
		hostInfo.OS,
		hostInfo.Architecture,
		hostInfo.Hostname,
	}, "|")
	hash := sha256.New()
	_, err := hash.Write([]byte(data))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate hash: %v", err)
	}
	fingerprint := hex.EncodeToString(hash.Sum(nil))
	return data, fingerprint, strings.Join(hostInfo.IPs, " - "), nil
}

func Permissions(targetPaths []string) (bool, error) {
	for _, path := range targetPaths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			// Path does not exist, attempt to create the directory
			err = os.MkdirAll(path, 0755)
			if err != nil {
				return false, fmt.Errorf("cannot create directory %s: %v", path, err)
			}
			// Directory created successfully, remove it
			defer os.Remove(path)
		} else if err != nil {
			// Error accessing the path
			return false, fmt.Errorf("error accessing path %s: %v", path, err)
		} else {
			if !info.IsDir() {
				return false, fmt.Errorf("path %s exists but is not a directory", path)
			}
			// Attempt to create a temporary file in the directory
			tmpFile, err := os.CreateTemp(path, "permtest_*")
			if err != nil {
				return false, fmt.Errorf("cannot create file in directory %s: %v", path, err)
			}
			// Close and remove the temporary file
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}
	}
	return true, nil
}
