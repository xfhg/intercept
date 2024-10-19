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

var remote_users = map[string]string{}

var filteredPolicies []Policy

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "(not final) Start the Remote Policy Execution endpoint",
	Long:  `(not final) Start the Remote Policy Execution endpoint with interactive interface for policy actions`,
	Run: func(cmd *cobra.Command, args []string) {
		//startSSHServer(filteredPolicies, outputDir)
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}

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
