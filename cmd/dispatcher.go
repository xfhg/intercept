package cmd

import (
	"fmt"
	"sync"

	"github.com/gookit/event"
)

var (
	dispatcher *Dispatcher
	once       sync.Once
)

// Dispatcher handles event dispatching and processing
type Dispatcher struct {
	manager *event.Manager
}

// GetDispatcher returns a singleton instance of Dispatcher
func GetDispatcher() *Dispatcher {
	once.Do(func() {
		dispatcher = &Dispatcher{
			manager: event.NewManager("intercept"),
		}
		dispatcher.registerEvents()
	})
	return dispatcher
}

// PolicyEventListener implements the event.Listener interface
type PolicyEventListener struct {
	handler func(e event.Event) error
}

func (l *PolicyEventListener) Handle(e event.Event) error {
	return l.handler(e)
}

// registerEvents sets up all the event listeners
func (d *Dispatcher) registerEvents() {
	d.manager.On("policy.scan", &PolicyEventListener{handler: d.handlePolicyScan}, event.Normal)
	d.manager.On("policy.assure", &PolicyEventListener{handler: d.handlePolicyAssure}, event.Normal)
	d.manager.On("policy.runtime", &PolicyEventListener{handler: d.handlePolicyRuntime}, event.Normal)
	d.manager.On("policy.api", &PolicyEventListener{handler: d.handlePolicyAPI}, event.Normal)
	d.manager.On("policy.yml", &PolicyEventListener{handler: d.handlePolicyYML}, event.Normal)
	d.manager.On("policy.toml", &PolicyEventListener{handler: d.handlePolicyTOML}, event.Normal)
	d.manager.On("policy.json", &PolicyEventListener{handler: d.handlePolicyJSON}, event.Normal)
	d.manager.On("policy.ini", &PolicyEventListener{handler: d.handlePolicyINI}, event.Normal)
	d.manager.On("policy.rego", &PolicyEventListener{handler: d.handlePolicyRego}, event.Normal)
}

// DispatchPolicyEvent dispatches a policy event based on its type
func (d *Dispatcher) DispatchPolicyEvent(policy Policy, targetDir string, filePaths []string) error {
	eventName := fmt.Sprintf("policy.%s", policy.Type)
	eventData := map[string]any{
		"policy":    policy,
		"targetDir": targetDir,
		"filePaths": filePaths,
	}

	// log.Debug().Msgf("Dispatching event: %s with data: %v", eventName, eventData)

	err, _ := d.manager.Fire(eventName, eventData)
	return err
}

// Event handlers
func (d *Dispatcher) handlePolicyScan(e event.Event) error {
	return processPolicyInWorker(e, "scan")
}

func (d *Dispatcher) handlePolicyAssure(e event.Event) error {
	return processPolicyInWorker(e, "assure")
}

func (d *Dispatcher) handlePolicyRuntime(e event.Event) error {
	return processPolicyInWorker(e, "runtime")
}

func (d *Dispatcher) handlePolicyAPI(e event.Event) error {
	return processPolicyInWorker(e, "api")
}

func (d *Dispatcher) handlePolicyYML(e event.Event) error {
	return processPolicyInWorker(e, "yml")
}

func (d *Dispatcher) handlePolicyTOML(e event.Event) error {
	return processPolicyInWorker(e, "toml")
}

func (d *Dispatcher) handlePolicyJSON(e event.Event) error {
	return processPolicyInWorker(e, "json")
}

func (d *Dispatcher) handlePolicyINI(e event.Event) error {
	return processPolicyInWorker(e, "ini")
}

func (d *Dispatcher) handlePolicyRego(e event.Event) error {
	return processPolicyInWorker(e, "rego")
}
