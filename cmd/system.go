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

var testEmbeddedCmd = &cobra.Command{
	Use:   "sys",
	Short: "Test intercept embedded core binaries",
	Long:  `This command extracts and runs the embedded rg and goss binaries to verify they are working correctly.`,
	Run: func(cmd *cobra.Command, args []string) {

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
func FingerprintHost(hostInfo *HostInfo) (string, string, error) {
	data := strings.Join([]string{
		hostInfo.MAC,
		hostInfo.OS,
		hostInfo.Architecture,
		hostInfo.Hostname,
	}, "|")
	hash := sha256.New()
	_, err := hash.Write([]byte(data))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate hash: %v", err)
	}
	fingerprint := hex.EncodeToString(hash.Sum(nil))
	return data, fingerprint, nil
}
