package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/segmentio/ksuid"
)

type HostInfo struct {
	Hostname     string
	OS           string
	Architecture string
	IPs          []string
	MAC          string
}

func watchPaths(paths ...string) {
	if len(paths) < 1 {
		log.Fatal().Msg("must specify at least one path to watch")
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Msgf("creating a new watcher: %s", err)
	}
	defer w.Close()

	// Start listening for events.
	go watchLoop(w, paths)

	// Add all paths from the commandline.
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			log.Fatal().Msgf("Failed to watch %q: %s", p, err)
		}
		log.Debug().Str("Observe", p).Msg("Watching Path")
	}

	log.Debug().Msg("Path Watcher Ready")
	<-make(chan struct{}) // Block forever
}

func watchLoop(w *fsnotify.Watcher, watchedPaths []string) {
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Error().Msgf("Watcher error: %s", err)

		case event, ok := <-w.Events:
			if !ok {
				return
			}

			log.Debug().Msgf("Watcher caught [%s] on [%s]", event.Op.String(), event.Name)

			// Process the event
			processEvent(event)

			// Re-add the watch for the file if it was removed or renamed
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				for _, path := range watchedPaths {
					if path == event.Name {
						// Wait a short time for the file to be recreated/renamed
						time.Sleep(100 * time.Millisecond)
						if err := w.Add(path); err != nil {
							log.Error().Msgf("Failed to re-add watch for %s: %s", path, err)
						} else {
							log.Debug().Msgf("Re-added watch for %s", path)
						}
						break
					}
				}
			}
		}
	}
}

func processEvent(e fsnotify.Event) {
	log.Debug().Str("fs", e.Name).Msg(e.String())

	policy, ok := LoadPolicyFromCache(e.Name)

	// Check if the watcher is targeting the directory
	if !ok {
		directoryCheck := GetDirectory(e.Name)
		log.Debug().Str("directory", directoryCheck).Msg(e.String())
		policy, ok = LoadPolicyFromCache(directoryCheck)
	}

	if ok {
		runID := fmt.Sprintf("%s-%s", ksuid.New().String(), NormalizeFilename(policy.ID))
		policy.RunID = runID
		log.Info().Str("policy", policy.ID).Str("runID", policy.RunID).Msgf("Triggered Policy run from watcher event [%s] ", e.Op.String())
		dispatcher.DispatchPolicyEvent(policy, targetDir, policy.Metadata.TargetInfo)
	} else {
		log.Error().Msgf("Policy not found in cache, watcher event [%s] didn't trigger policy process for: %s", e.Op.String(), e.Name)
	}
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
