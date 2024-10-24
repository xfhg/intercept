package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bwish "github.com/charmbracelet/wish/bubbletea"
	"github.com/pkg/exec"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

// this is for --remote
const (
	host = "0.0.0.0"
	port = "23234"
)

var remote_users = map[string]string{}
var filteredPolicies []Policy

// var remoteCmd = &cobra.Command{
// 	Use:   "remote",
// 	Short: "(not final) Load the Remote Policy Execution",
// 	Long:  `(not final) Load the Remote Policy Execution endpoint with interactive interface for policy actions`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		log.Fatal().Msg("Not yet implemented, use `observe --remote` to start the remote policy execution interface")
// 	},
// }

var (
	remoteFlags struct {
		user               string
		password           string
		askPass            bool
		identityFile       string
		port               int
		sudo               bool
		asUser             string
		concurrency        int
		destPath           string
		files              []string
		execute            string
		force              bool
		zip                bool
		gosshPath          string
		inventory          string
		configFile         string
		passFile           string
		passphrase         string
		vaultPassFile      string
		listHosts          bool
		proxyServer        string
		proxyPort          int
		proxyUser          string
		proxyPassword      string
		proxyIdentityFiles string
		proxyPassphrase    string
		commandTimeout     int
		taskTimeout        int
		connTimeout        int
		lang               string
		commandBlacklist   []string
		remove             bool
		tmpDir             string
	}

	remoteCmd = &cobra.Command{
		Use:   "remote",
		Short: "Execute remote operations using gossh",
		Long:  `Execute remote operations like run, push, fetch, and script using embedded gossh`,
	}

	runCmd = &cobra.Command{
		Use:   "run [HOST...]",
		Short: "Execute commands on target hosts",
		Long: `Execute commands on target hosts.

Examples:
  Execute command 'uptime' on target hosts:
  $ intercept remote run host1 host2 --r-execute "uptime" --r-auth.user zhangsan --r-auth.ask-pass

  Use sudo as root to execute command on target hosts:
  $ intercept remote run host[1-2] --r-execute "uptime" --r-auth.user zhangsan --r-run.sudo

  Use sudo as other user 'mysql' to execute command on target hosts:
  $ intercept remote run host[1-2] --r-execute "uptime" --r-auth.user zhangsan --r-run.sudo --r-run.as-user mysql`,
		RunE: executeRun,
	}

	pushCmd = &cobra.Command{
		Use:   "push [HOST...]",
		Short: "Copy local files and dirs to target hosts",
		RunE:  executePush,
	}

	fetchCmd = &cobra.Command{
		Use:   "fetch [HOST...]",
		Short: "Copy files from target hosts to local",
		RunE:  executeFetch,
	}

	scriptCmd = &cobra.Command{
		Use:   "script [HOST...]",
		Short: "Execute a local shell script on target hosts",
		RunE:  executeScript,
	}
)

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(runCmd, pushCmd, fetchCmd, scriptCmd)

	// Authentication flags
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.user, "r-auth.user", os.Getenv("USER"), "login user")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.password, "r-auth.password", "", "password of login user")
	remoteCmd.PersistentFlags().BoolVar(&remoteFlags.askPass, "r-auth.ask-pass", false, "ask for password")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.passFile, "r-auth.pass-file", "", "file that holds the password of login user")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.identityFile, "r-auth.identity-files", "", "identity files")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.passphrase, "r-auth.passphrase", "", "passphrase of the identity files")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.vaultPassFile, "r-auth.vault-pass-file", "", "vault password file")

	// Host flags
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.inventory, "r-hosts.inventory", "", "file that holds the target hosts")
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.port, "r-hosts.port", 22, "port of the target hosts")
	remoteCmd.PersistentFlags().BoolVar(&remoteFlags.listHosts, "r-hosts.list", false, "outputs a list of target hosts")

	// Run flags
	remoteCmd.PersistentFlags().BoolVar(&remoteFlags.sudo, "r-run.sudo", false, "use sudo")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.asUser, "r-run.as-user", "root", "run as user")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.lang, "r-run.lang", "", "specify i18n while executing command")
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.concurrency, "r-run.concurrency", 10, "number of concurrent connections")
	remoteCmd.PersistentFlags().StringSliceVar(&remoteFlags.commandBlacklist, "r-run.command-blacklist", []string{}, "commands that are prohibited")

	// Proxy flags
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.proxyServer, "r-proxy.server", "", "proxy server address")
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.proxyPort, "r-proxy.port", 22, "proxy server port")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.proxyUser, "r-proxy.user", "", "login user for proxy")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.proxyPassword, "r-proxy.password", "", "password for proxy")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.proxyIdentityFiles, "r-proxy.identity-files", "", "identity files for proxy")
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.proxyPassphrase, "r-proxy.passphrase", "", "passphrase of the identity files for proxy")

	// Timeout flags
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.commandTimeout, "r-timeout.command", 0, "timeout for each command")
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.taskTimeout, "r-timeout.task", 0, "timeout for the entire task")
	remoteCmd.PersistentFlags().IntVar(&remoteFlags.connTimeout, "r-timeout.conn", 10, "timeout for connecting each host")

	// Config flag
	remoteCmd.PersistentFlags().StringVar(&remoteFlags.configFile, "r-config", "", "remote config file")

	// Add command-specific flags for runCmd
	runCmd.Flags().StringVar(&remoteFlags.execute, "r-execute", "", "commands to be executed on target hosts")
	runCmd.Flags().Bool("r-no-safe-check", false, "ignore dangerous commands (from '--r-run.command-blacklist') check")
	// Mark required flags
	runCmd.MarkFlagRequired("r-execute")

	// Add command-specific flags for scriptCmd
	scriptCmd.Flags().StringVar(&remoteFlags.execute, "r-script", "", "a shell script to be executed on target hosts")
	scriptCmd.Flags().StringVar(&remoteFlags.destPath, "r-dest-path", "/tmp", "path of target hosts where the script will be copied to")
	scriptCmd.Flags().BoolVar(&remoteFlags.force, "r-force", false, "allow overwrite script file if it already exists on target hosts")
	scriptCmd.Flags().BoolVar(&remoteFlags.remove, "r-remove", false, "remove the copied script after execution")
	// Mark required flags
	scriptCmd.MarkFlagRequired("r-script")

	// Add command-specific flags for pushCmd
	pushCmd.Flags().StringSliceVar(&remoteFlags.files, "r-files", []string{}, "local files/dirs to be copied to target hosts")
	pushCmd.Flags().StringVar(&remoteFlags.destPath, "r-dest-path", "/tmp", "path of target hosts where files/dirs will be copied to")
	pushCmd.Flags().BoolVar(&remoteFlags.force, "r-force", false, "allow overwrite files/dirs if they already exist on target hosts")
	pushCmd.Flags().BoolVar(&remoteFlags.zip, "r-zip", false, "enable zip files ('unzip' must be installed on target hosts)")

	// Mark required flags
	pushCmd.MarkFlagRequired("r-files")

	// Add command-specific flags for fetchCmd
	fetchCmd.Flags().StringSliceVar(&remoteFlags.files, "r-files", []string{}, "files/dirs on target hosts that to be copied")
	fetchCmd.Flags().StringVar(&remoteFlags.destPath, "r-dest-path", "", "local directory that files/dirs from target hosts will be copied to")
	fetchCmd.Flags().BoolVar(&remoteFlags.zip, "r-zip", false, "enable zip files ('zip' must be installed on target hosts)")
	fetchCmd.Flags().StringVar(&remoteFlags.tmpDir, "r-tmp-dir", "$HOME", "directory for storing temporary zip file on target hosts, only useful if the -z flag is used")

	// Mark required flags
	fetchCmd.MarkFlagRequired("r-files")
	fetchCmd.MarkFlagRequired("r-dest-path")

	if runtime.GOOS == "windows" || (runtime.GOOS == "linux" && runtime.GOARCH == "arm") {
		log.Fatal().Msg("INTERCEPT REMOTE currently not supported on your architecture")
	} else {

		gosshPath, err := prepareGosshExecutable()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to prepare gossh binary")
		}
		remoteFlags.gosshPath = gosshPath
	}
}

func executeRun(cmd *cobra.Command, args []string) error {
	// Add validation for gosshPath
	if remoteFlags.gosshPath == "" {
		return fmt.Errorf("gossh binary path not set")
	}

	// Modified validation: allow either direct hosts or inventory file
	if len(args) == 0 && remoteFlags.inventory == "" {
		return fmt.Errorf("either hosts or --r-hosts.inventory flag is required")
	}
	if remoteFlags.execute == "" {
		return fmt.Errorf("--r-execute flag is required")
	}

	// Debug log the gossh path
	log.Debug().Str("gosshPath", remoteFlags.gosshPath).Msg("Using gossh binary")

	// Get the no-safe-check flag
	noSafeCheck, _ := cmd.Flags().GetBool("r-no-safe-check")

	// Prepare the gossh command arguments
	gosshArgs := []string{"command"}

	// Add hosts from args if provided
	if len(args) > 0 {
		gosshArgs = append(gosshArgs, args...)
	}

	// Add inventory file if provided
	if remoteFlags.inventory != "" {
		gosshArgs = append(gosshArgs, "--hosts.inventory", remoteFlags.inventory)
	}

	// Add config file if provided
	if remoteFlags.configFile != "" {
		gosshArgs = append(gosshArgs, "--config", remoteFlags.configFile)
	}

	// Add the execute command
	gosshArgs = append(gosshArgs, "--execute", remoteFlags.execute)

	// Add no-safe-check if enabled
	if noSafeCheck {
		gosshArgs = append(gosshArgs, "--no-safe-check")
	}

	// Add authentication flags
	if remoteFlags.user != os.Getenv("USER") {
		gosshArgs = append(gosshArgs, "--auth.user", remoteFlags.user)
	}
	if remoteFlags.password != "" {
		gosshArgs = append(gosshArgs, "--auth.password", remoteFlags.password)
	}
	if remoteFlags.askPass {
		gosshArgs = append(gosshArgs, "--auth.ask-pass")
	}
	if remoteFlags.passFile != "" {
		gosshArgs = append(gosshArgs, "--auth.pass-file", remoteFlags.passFile)
	}
	if remoteFlags.identityFile != "" {
		gosshArgs = append(gosshArgs, "--auth.identity-files", remoteFlags.identityFile)
	}
	if remoteFlags.passphrase != "" {
		gosshArgs = append(gosshArgs, "--auth.passphrase", remoteFlags.passphrase)
	}
	if remoteFlags.vaultPassFile != "" {
		gosshArgs = append(gosshArgs, "--auth.vault-pass-file", remoteFlags.vaultPassFile)
	}

	// Add host flags
	if remoteFlags.port != 22 {
		gosshArgs = append(gosshArgs, "--hosts.port", fmt.Sprintf("%d", remoteFlags.port))
	}
	if remoteFlags.listHosts {
		gosshArgs = append(gosshArgs, "--hosts.list")
	}

	// Add run flags
	if remoteFlags.sudo {
		gosshArgs = append(gosshArgs, "--run.sudo")
	}
	if remoteFlags.asUser != "root" {
		gosshArgs = append(gosshArgs, "--run.as-user", remoteFlags.asUser)
	}
	if remoteFlags.lang != "" {
		gosshArgs = append(gosshArgs, "--run.lang", remoteFlags.lang)
	}
	if remoteFlags.concurrency != 1 {
		gosshArgs = append(gosshArgs, "--run.concurrency", fmt.Sprintf("%d", remoteFlags.concurrency))
	}
	if len(remoteFlags.commandBlacklist) > 0 && !noSafeCheck {
		gosshArgs = append(gosshArgs, "--run.command-blacklist", strings.Join(remoteFlags.commandBlacklist, ","))
	}

	// Add proxy flags
	if remoteFlags.proxyServer != "" {
		gosshArgs = append(gosshArgs, "--proxy.server", remoteFlags.proxyServer)
	}
	if remoteFlags.proxyPort != 22 {
		gosshArgs = append(gosshArgs, "--proxy.port", fmt.Sprintf("%d", remoteFlags.proxyPort))
	}
	if remoteFlags.proxyUser != "" {
		gosshArgs = append(gosshArgs, "--proxy.user", remoteFlags.proxyUser)
	}
	if remoteFlags.proxyPassword != "" {
		gosshArgs = append(gosshArgs, "--proxy.password", remoteFlags.proxyPassword)
	}
	if remoteFlags.proxyIdentityFiles != "" {
		gosshArgs = append(gosshArgs, "--proxy.identity-files", remoteFlags.proxyIdentityFiles)
	}
	if remoteFlags.proxyPassphrase != "" {
		gosshArgs = append(gosshArgs, "--proxy.passphrase", remoteFlags.proxyPassphrase)
	}

	// Add timeout flags
	if remoteFlags.commandTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.command", fmt.Sprintf("%d", remoteFlags.commandTimeout))
	}
	if remoteFlags.taskTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.task", fmt.Sprintf("%d", remoteFlags.taskTimeout))
	}
	if remoteFlags.connTimeout != 10 {
		gosshArgs = append(gosshArgs, "--timeout.conn", fmt.Sprintf("%d", remoteFlags.connTimeout))
	}

	//append output type
	if verbosity < 2 {
		gosshArgs = append(gosshArgs, "--output.json")
	}
	// Execute gossh command
	execCmd := exec.Command(remoteFlags.gosshPath, gosshArgs...)

	// Connect stdout and stderr
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command
	log.Debug().
		Str("command", remoteFlags.gosshPath).
		Strs("args", gosshArgs).
		Msg("Executing remote command")

	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute remote command: %w", err)
	}

	return nil
}

func executePush(cmd *cobra.Command, args []string) error {
	// Add validation for gosshPath
	if remoteFlags.gosshPath == "" {
		return fmt.Errorf("gossh binary path not set")
	}

	// Modified validation: allow either direct hosts or inventory file
	if len(args) == 0 && remoteFlags.inventory == "" {
		return fmt.Errorf("either hosts or --r-hosts.inventory flag is required")
	}
	if remoteFlags.execute == "" {
		return fmt.Errorf("--r-execute flag is required")
	}

	// Debug log the gossh path
	log.Debug().Str("gosshPath", remoteFlags.gosshPath).Msg("Using gossh binary")

	// Prepare the gossh command arguments
	gosshArgs := []string{"push"}

	// Add hosts from args if provided
	if len(args) > 0 {
		gosshArgs = append(gosshArgs, args...)
	}

	// Add inventory file if provided
	if remoteFlags.inventory != "" {
		gosshArgs = append(gosshArgs, "--hosts.inventory", remoteFlags.inventory)
	}

	// Add files to be copied
	gosshArgs = append(gosshArgs, "--files", strings.Join(remoteFlags.files, ","))

	// Add destination path
	gosshArgs = append(gosshArgs, "--dest-path", remoteFlags.destPath)

	// Add force flag if enabled
	if remoteFlags.force {
		gosshArgs = append(gosshArgs, "--force")
	}

	// Add zip flag if enabled
	if remoteFlags.zip {
		gosshArgs = append(gosshArgs, "--zip")
	}

	// Add authentication flags
	if remoteFlags.user != os.Getenv("USER") {
		gosshArgs = append(gosshArgs, "--auth.user", remoteFlags.user)
	}
	if remoteFlags.password != "" {
		gosshArgs = append(gosshArgs, "--auth.password", remoteFlags.password)
	}
	if remoteFlags.askPass {
		gosshArgs = append(gosshArgs, "--auth.ask-pass")
	}
	if remoteFlags.passFile != "" {
		gosshArgs = append(gosshArgs, "--auth.pass-file", remoteFlags.passFile)
	}
	if remoteFlags.identityFile != "" {
		gosshArgs = append(gosshArgs, "--auth.identity-files", remoteFlags.identityFile)
	}
	if remoteFlags.passphrase != "" {
		gosshArgs = append(gosshArgs, "--auth.passphrase", remoteFlags.passphrase)
	}
	if remoteFlags.vaultPassFile != "" {
		gosshArgs = append(gosshArgs, "--auth.vault-pass-file", remoteFlags.vaultPassFile)
	}

	// Add host flags
	if remoteFlags.port != 22 {
		gosshArgs = append(gosshArgs, "--hosts.port", fmt.Sprintf("%d", remoteFlags.port))
	}
	if remoteFlags.listHosts {
		gosshArgs = append(gosshArgs, "--hosts.list")
	}

	// Add run flags
	if remoteFlags.sudo {
		gosshArgs = append(gosshArgs, "--run.sudo")
	}
	if remoteFlags.asUser != "root" {
		gosshArgs = append(gosshArgs, "--run.as-user", remoteFlags.asUser)
	}
	if remoteFlags.lang != "" {
		gosshArgs = append(gosshArgs, "--run.lang", remoteFlags.lang)
	}
	if remoteFlags.concurrency != 1 {
		gosshArgs = append(gosshArgs, "--run.concurrency", fmt.Sprintf("%d", remoteFlags.concurrency))
	}
	if len(remoteFlags.commandBlacklist) > 0 {
		gosshArgs = append(gosshArgs, "--run.command-blacklist", strings.Join(remoteFlags.commandBlacklist, ","))
	}

	// Add proxy flags
	if remoteFlags.proxyServer != "" {
		gosshArgs = append(gosshArgs, "--proxy.server", remoteFlags.proxyServer)
	}
	if remoteFlags.proxyPort != 22 {
		gosshArgs = append(gosshArgs, "--proxy.port", fmt.Sprintf("%d", remoteFlags.proxyPort))
	}
	if remoteFlags.proxyUser != "" {
		gosshArgs = append(gosshArgs, "--proxy.user", remoteFlags.proxyUser)
	}
	if remoteFlags.proxyPassword != "" {
		gosshArgs = append(gosshArgs, "--proxy.password", remoteFlags.proxyPassword)
	}
	if remoteFlags.proxyIdentityFiles != "" {
		gosshArgs = append(gosshArgs, "--proxy.identity-files", remoteFlags.proxyIdentityFiles)
	}
	if remoteFlags.proxyPassphrase != "" {
		gosshArgs = append(gosshArgs, "--proxy.passphrase", remoteFlags.proxyPassphrase)
	}

	// Add timeout flags
	if remoteFlags.commandTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.command", fmt.Sprintf("%d", remoteFlags.commandTimeout))
	}
	if remoteFlags.taskTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.task", fmt.Sprintf("%d", remoteFlags.taskTimeout))
	}
	if remoteFlags.connTimeout != 10 {
		gosshArgs = append(gosshArgs, "--timeout.conn", fmt.Sprintf("%d", remoteFlags.connTimeout))
	}

	//append output type
	if verbosity < 2 {
		gosshArgs = append(gosshArgs, "--output.json")
	}

	// Execute gossh command
	execCmd := exec.Command(remoteFlags.gosshPath, gosshArgs...)

	// Connect stdout and stderr
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command
	log.Debug().
		Str("command", remoteFlags.gosshPath).
		Strs("args", gosshArgs).
		Msg("Executing push command")

	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute push command: %w", err)
	}

	return nil
}

func executeFetch(cmd *cobra.Command, args []string) error {
	// Add validation for gosshPath
	if remoteFlags.gosshPath == "" {
		return fmt.Errorf("gossh binary path not set")
	}

	// Modified validation: allow either direct hosts or inventory file
	if len(args) == 0 && remoteFlags.inventory == "" {
		return fmt.Errorf("either hosts or --r-hosts.inventory flag is required")
	}
	if len(remoteFlags.files) == 0 {
		return fmt.Errorf("--r-files flag is required")
	}
	if remoteFlags.destPath == "" {
		return fmt.Errorf("--r-dest-path flag is required")
	}

	// Debug log the gossh path
	log.Debug().Str("gosshPath", remoteFlags.gosshPath).Msg("Using gossh binary")

	// Prepare the gossh command arguments
	gosshArgs := []string{"fetch"}

	// Add hosts from args if provided
	if len(args) > 0 {
		gosshArgs = append(gosshArgs, args...)
	}

	// Add inventory file if provided
	if remoteFlags.inventory != "" {
		gosshArgs = append(gosshArgs, "--hosts.inventory", remoteFlags.inventory)
	}

	// Add files to be copied
	gosshArgs = append(gosshArgs, "--files", strings.Join(remoteFlags.files, ","))

	// Add destination path
	gosshArgs = append(gosshArgs, "--dest-path", remoteFlags.destPath)

	// Add zip flag if enabled
	if remoteFlags.zip {
		gosshArgs = append(gosshArgs, "--zip")
	}

	// Add temporary directory for zip files if zip is enabled
	if remoteFlags.zip && remoteFlags.tmpDir != "" {
		gosshArgs = append(gosshArgs, "--tmp-dir", remoteFlags.tmpDir)
	}

	// Add authentication flags
	if remoteFlags.user != os.Getenv("USER") {
		gosshArgs = append(gosshArgs, "--auth.user", remoteFlags.user)
	}
	if remoteFlags.password != "" {
		gosshArgs = append(gosshArgs, "--auth.password", remoteFlags.password)
	}
	if remoteFlags.askPass {
		gosshArgs = append(gosshArgs, "--auth.ask-pass")
	}
	if remoteFlags.passFile != "" {
		gosshArgs = append(gosshArgs, "--auth.pass-file", remoteFlags.passFile)
	}
	if remoteFlags.identityFile != "" {
		gosshArgs = append(gosshArgs, "--auth.identity-files", remoteFlags.identityFile)
	}
	if remoteFlags.passphrase != "" {
		gosshArgs = append(gosshArgs, "--auth.passphrase", remoteFlags.passphrase)
	}
	if remoteFlags.vaultPassFile != "" {
		gosshArgs = append(gosshArgs, "--auth.vault-pass-file", remoteFlags.vaultPassFile)
	}

	// Add host flags
	if remoteFlags.port != 22 {
		gosshArgs = append(gosshArgs, "--hosts.port", fmt.Sprintf("%d", remoteFlags.port))
	}
	if remoteFlags.listHosts {
		gosshArgs = append(gosshArgs, "--hosts.list")
	}

	// Add run flags
	if remoteFlags.sudo {
		gosshArgs = append(gosshArgs, "--run.sudo")
	}
	if remoteFlags.asUser != "root" {
		gosshArgs = append(gosshArgs, "--run.as-user", remoteFlags.asUser)
	}
	if remoteFlags.lang != "" {
		gosshArgs = append(gosshArgs, "--run.lang", remoteFlags.lang)
	}
	if remoteFlags.concurrency != 1 {
		gosshArgs = append(gosshArgs, "--run.concurrency", fmt.Sprintf("%d", remoteFlags.concurrency))
	}
	if len(remoteFlags.commandBlacklist) > 0 {
		gosshArgs = append(gosshArgs, "--run.command-blacklist", strings.Join(remoteFlags.commandBlacklist, ","))
	}

	// Add proxy flags
	if remoteFlags.proxyServer != "" {
		gosshArgs = append(gosshArgs, "--proxy.server", remoteFlags.proxyServer)
	}
	if remoteFlags.proxyPort != 22 {
		gosshArgs = append(gosshArgs, "--proxy.port", fmt.Sprintf("%d", remoteFlags.proxyPort))
	}
	if remoteFlags.proxyUser != "" {
		gosshArgs = append(gosshArgs, "--proxy.user", remoteFlags.proxyUser)
	}
	if remoteFlags.proxyPassword != "" {
		gosshArgs = append(gosshArgs, "--proxy.password", remoteFlags.proxyPassword)
	}
	if remoteFlags.proxyIdentityFiles != "" {
		gosshArgs = append(gosshArgs, "--proxy.identity-files", remoteFlags.proxyIdentityFiles)
	}
	if remoteFlags.proxyPassphrase != "" {
		gosshArgs = append(gosshArgs, "--proxy.passphrase", remoteFlags.proxyPassphrase)
	}

	// Add timeout flags
	if remoteFlags.commandTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.command", fmt.Sprintf("%d", remoteFlags.commandTimeout))
	}
	if remoteFlags.taskTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.task", fmt.Sprintf("%d", remoteFlags.taskTimeout))
	}
	if remoteFlags.connTimeout != 10 {
		gosshArgs = append(gosshArgs, "--timeout.conn", fmt.Sprintf("%d", remoteFlags.connTimeout))
	}

	//append output type
	if verbosity < 2 {
		gosshArgs = append(gosshArgs, "--output.json")
	}

	// Execute gossh command
	execCmd := exec.Command(remoteFlags.gosshPath, gosshArgs...)

	// Connect stdout and stderr
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command
	log.Debug().
		Str("command", remoteFlags.gosshPath).
		Strs("args", gosshArgs).
		Msg("Executing fetch command")

	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute fetch command: %w", err)
	}

	return nil
}

func executeScript(cmd *cobra.Command, args []string) error {
	// Add validation for gosshPath
	if remoteFlags.gosshPath == "" {
		return fmt.Errorf("gossh binary path not set")
	}

	// Modified validation: allow either direct hosts or inventory file
	if len(args) == 0 && remoteFlags.inventory == "" {
		return fmt.Errorf("either hosts or --r-hosts.inventory flag is required")
	}
	if remoteFlags.execute == "" {
		return fmt.Errorf("--r-execute flag is required")
	}

	// Debug log the gossh path
	log.Debug().Str("gosshPath", remoteFlags.gosshPath).Msg("Using gossh binary")

	// Get the no-safe-check flag
	noSafeCheck, _ := cmd.Flags().GetBool("r-no-safe-check")

	// Prepare the gossh command arguments
	gosshArgs := []string{"script"}

	// Add hosts from args if provided
	if len(args) > 0 {
		gosshArgs = append(gosshArgs, args...)
	}

	// Add inventory file if provided
	if remoteFlags.inventory != "" {
		gosshArgs = append(gosshArgs, "--hosts.inventory", remoteFlags.inventory)
	}

	// Add the execute command
	gosshArgs = append(gosshArgs, "--execute", remoteFlags.execute)

	// Add destination path
	gosshArgs = append(gosshArgs, "--dest-path", remoteFlags.destPath)

	// Add force flag if enabled
	if remoteFlags.force {
		gosshArgs = append(gosshArgs, "--force")
	}

	// Add remove flag if enabled
	if remoteFlags.remove {
		gosshArgs = append(gosshArgs, "--remove")
	}
	// Add no-safe-check if enabled
	if noSafeCheck {
		gosshArgs = append(gosshArgs, "--no-safe-check")
	}

	// Add authentication flags
	if remoteFlags.user != os.Getenv("USER") {
		gosshArgs = append(gosshArgs, "--auth.user", remoteFlags.user)
	}
	if remoteFlags.password != "" {
		gosshArgs = append(gosshArgs, "--auth.password", remoteFlags.password)
	}
	if remoteFlags.askPass {
		gosshArgs = append(gosshArgs, "--auth.ask-pass")
	}
	if remoteFlags.passFile != "" {
		gosshArgs = append(gosshArgs, "--auth.pass-file", remoteFlags.passFile)
	}
	if remoteFlags.identityFile != "" {
		gosshArgs = append(gosshArgs, "--auth.identity-files", remoteFlags.identityFile)
	}
	if remoteFlags.passphrase != "" {
		gosshArgs = append(gosshArgs, "--auth.passphrase", remoteFlags.passphrase)
	}
	if remoteFlags.vaultPassFile != "" {
		gosshArgs = append(gosshArgs, "--auth.vault-pass-file", remoteFlags.vaultPassFile)
	}

	// Add host flags
	if remoteFlags.port != 22 {
		gosshArgs = append(gosshArgs, "--hosts.port", fmt.Sprintf("%d", remoteFlags.port))
	}
	if remoteFlags.listHosts {
		gosshArgs = append(gosshArgs, "--hosts.list")
	}

	// Add run flags
	if remoteFlags.sudo {
		gosshArgs = append(gosshArgs, "--run.sudo")
	}
	if remoteFlags.asUser != "root" {
		gosshArgs = append(gosshArgs, "--run.as-user", remoteFlags.asUser)
	}
	if remoteFlags.lang != "" {
		gosshArgs = append(gosshArgs, "--run.lang", remoteFlags.lang)
	}
	if remoteFlags.concurrency != 1 {
		gosshArgs = append(gosshArgs, "--run.concurrency", fmt.Sprintf("%d", remoteFlags.concurrency))
	}
	if len(remoteFlags.commandBlacklist) > 0 && !noSafeCheck {
		gosshArgs = append(gosshArgs, "--run.command-blacklist", strings.Join(remoteFlags.commandBlacklist, ","))
	}

	// Add proxy flags
	if remoteFlags.proxyServer != "" {
		gosshArgs = append(gosshArgs, "--proxy.server", remoteFlags.proxyServer)
	}
	if remoteFlags.proxyPort != 22 {
		gosshArgs = append(gosshArgs, "--proxy.port", fmt.Sprintf("%d", remoteFlags.proxyPort))
	}
	if remoteFlags.proxyUser != "" {
		gosshArgs = append(gosshArgs, "--proxy.user", remoteFlags.proxyUser)
	}
	if remoteFlags.proxyPassword != "" {
		gosshArgs = append(gosshArgs, "--proxy.password", remoteFlags.proxyPassword)
	}
	if remoteFlags.proxyIdentityFiles != "" {
		gosshArgs = append(gosshArgs, "--proxy.identity-files", remoteFlags.proxyIdentityFiles)
	}
	if remoteFlags.proxyPassphrase != "" {
		gosshArgs = append(gosshArgs, "--proxy.passphrase", remoteFlags.proxyPassphrase)
	}

	// Add timeout flags
	if remoteFlags.commandTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.command", fmt.Sprintf("%d", remoteFlags.commandTimeout))
	}
	if remoteFlags.taskTimeout != 0 {
		gosshArgs = append(gosshArgs, "--timeout.task", fmt.Sprintf("%d", remoteFlags.taskTimeout))
	}
	if remoteFlags.connTimeout != 10 {
		gosshArgs = append(gosshArgs, "--timeout.conn", fmt.Sprintf("%d", remoteFlags.connTimeout))
	}

	//append output type
	if verbosity < 2 {
		gosshArgs = append(gosshArgs, "--output.json")
	}

	// Execute gossh command
	execCmd := exec.Command(remoteFlags.gosshPath, gosshArgs...)

	// Connect stdout and stderr
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Run the command
	log.Debug().
		Str("command", remoteFlags.gosshPath).
		Strs("args", gosshArgs).
		Msg("Executing script command")

	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute script command: %w", err)
	}

	return nil
}

// ----------------------------------------------------------------------------------
// Remote Policy Execution
// ----------------------------------------------------------------------------------
// this is for --remote

func authenticatedBubbleteaMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			for name, pubkey := range remote_users {
				parsed, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(pubkey))
				if ssh.KeysEqual(s.PublicKey(), parsed) {
					wish.Println(s, fmt.Sprintf("┗━━━┫ Authenticated as %s \n\n", name))
					bwish.Middleware(policyActionHandler)(next)(s)
					return
				}
			}
			wish.Println(s, "┗━━━┫ Authentication failed ╳ \n\n")
			s.Close()
		}
	}
}

func policyActionHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return newModel(), bwish.MakeOptions(s)
}

// Model represents the state of our Bubble Tea program
type model struct {
	choices  []policyChoice
	cursor   int
	selected map[int]struct{}
	message  string
}

type policyChoice struct {
	policy Policy
}

func newModel() *model {
	choices := make([]policyChoice, len(filteredPolicies))
	for i, policy := range filteredPolicies {
		choices[i] = policyChoice{
			policy: policy,
		}
	}
	return &model{
		choices:  choices,
		selected: make(map[int]struct{}),
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case policyExecutedMsg:
		m.message = string(msg)
		// Clear selected policies after execution
		m.selected = make(map[int]struct{})
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "r":
			return m, m.runSelectedPolicies
		}
	}
	return m, nil
}

func (m *model) View() string {
	s := "Select policies to run (use arrow keys, space to select, 'r' to run):\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		// Get the status of the policy
		status, ok := loadResultFromCache(choice.policy.ID)

		if !ok {
			status = "⚪ Not executed"
		}

		s += fmt.Sprintf("%s [%s] %s \t - %s \n", cursor, checked, choice.policy.ID, status)
	}

	if m.message != "" {
		s += "\n" + m.message + "\n"
	}

	s += "\nPress q to quit.\n"

	return s
}

func (m *model) runSelectedPolicies() tea.Msg {
	dispatcher := GetDispatcher()
	var messages []string

	for i := range m.selected {
		policy := m.choices[i].policy

		runID := fmt.Sprintf("%s-%s", ksuid.New().String(), NormalizeFilename(policy.ID))
		log.Info().Str("policy", policy.ID).Str("runID", runID).Msg("Executing policy via REMOTE CALL")

		// Update the RunID for the policy in the model
		policy.RunID = runID
		m.choices[i].policy.RunID = runID

		err := dispatcher.DispatchPolicyEvent(policy, "", nil)
		timestamp := time.Now().Format("15:04:05")
		if err != nil {
			log.Error().Err(err).Str("policy", policy.ID).Str("runID", runID).Msg("Failed to execute policy via REMOTE CALL")
			messages = append(messages, fmt.Sprintf("[%s] Failed to execute policy %s: %v", timestamp, policy.ID, err))
		} else {
			log.Info().Str("policy", policy.ID).Str("runID", runID).Msg("Successfully executed policy via REMOTE CALL")
			messages = append(messages, fmt.Sprintf("[%s] Policy %s executed successfully", timestamp, policy.ID))
		}
	}

	if len(messages) > 0 {
		return policyExecutedMsg(strings.Join(messages, "\n"))
	}

	return nil
}

func startSSHServer(policies []Policy, outputDir string) error {
	// Filter policies to include only those with Type == "runtime"
	var runtimePolicies []Policy
	for _, policy := range policies {
		if policy.Type == "runtime" {
			runtimePolicies = append(runtimePolicies, policy)
		}
	}
	filteredPolicies = runtimePolicies

	hostKeyPath := filepath.Join(outputDir, "_rpe/id_ed25519")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithBannerHandler(func(ctx ssh.Context) string {
			return "\n\n┏━ INTERCEPT Remote Policy Execution Endpoint\n┃\n"
		}),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			authenticatedBubbleteaMiddleware(),
			// logging.Middleware(),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	log.Info().Str("host", host).Str("port", port).Msg("INTERCEPT Remote Policy Execution Endpoint")
	if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return fmt.Errorf("could not start server: %w", err)
	}

	return nil
}

type policyExecutedMsg string

func authKeysToMap(arr []string) map[string]string {
	users := make(map[string]string)
	for _, s := range arr {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) == 2 {
			alias := parts[0]
			sshKey := parts[1]
			users[alias] = sshKey
		} else {
			log.Error().Msgf("Invalid key format for entry: %s\n", s)
		}
	}
	return users
}
