package cmd

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bwish "github.com/charmbracelet/wish/bubbletea"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

const (
	host = "localhost"
	port = "23234"
)

var users = map[string]string{

	"Flavio": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICtFRLdSvayFQwQdIOk6NKuEpEK7KvYBQz8LUVerSo8T",
	"Arka":   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJyubt40tutUSi3FQqcEzbDUu14RdLstEbURvX/M2bM/",
	// You can add add your name and public key here :)
}

var filteredPolicies []Policy

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "(not final) Start the SSH server",
	Long:  `(not final) Start the SSH server with interactive interface for policy actions`,
	Run: func(cmd *cobra.Command, args []string) {
		startSSHServer(filteredPolicies, outputDir)
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}

func authenticatedBubbleteaMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			for name, pubkey := range users {
				parsed, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(pubkey))
				if ssh.KeysEqual(s.PublicKey(), parsed) {
					wish.Println(s, fmt.Sprintf("-/- Welcome, %s! -/- \n\n", name))
					bwish.Middleware(policyActionHandler)(next)(s)
					return
				}
			}
			wish.Println(s, "Authentication failed. Goodbye!")
			s.Close()
		}
	}
}

func policyActionHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return newModel(), bwish.MakeOptions(s)
}

// Model represents the state of our Bubble Tea program
type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	message  string
}

func newModel() model {
	choices := make([]string, len(filteredPolicies))
	for i, policy := range filteredPolicies {
		choices[i] = "policy." + policy.ID
	}
	return model{
		choices:  choices,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
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

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	if m.message != "" {
		s += "\n" + m.message + "\n"
	}

	s += "\nPress q to quit.\n"

	return s
}

func (m model) runSelectedPolicies() tea.Msg {
	dispatcher := GetDispatcher()
	var messages []string

	for i := range m.selected {
		policy := filteredPolicies[i]

		runID := fmt.Sprintf("%s-%s", ksuid.New().String(), NormalizeFilename(policy.ID))
		log.Info().Str("policy", policy.ID).Str("runID", runID).Msg("Executing policy via REMOTE CALL")

		// Set the RunID for the policy
		policy.RunID = runID

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
	filteredPolicies = policies

	hostKeyPath := filepath.Join(outputDir, "_rpe/id_ed25519")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithBannerHandler(func(ctx ssh.Context) string {
			return "\n\n-/- INTERCEPT Remote Policy Execution Endpoint -/-\n\n\n"
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
