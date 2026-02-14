package cmd

import "sync"

var (
	statusHandlerMu     sync.RWMutex
	policyStatusHandler = func(Policy, bool) {}
)

// SetPolicyStatusHandler registers a handler invoked when an assure policy finishes executing.
// Passing nil resets the handler to a no-op implementation.
func SetPolicyStatusHandler(handler func(Policy, bool)) {
	statusHandlerMu.Lock()
	defer statusHandlerMu.Unlock()

	if handler == nil {
		policyStatusHandler = func(Policy, bool) {}
		return
	}

	policyStatusHandler = handler
}

func reportPolicyStatus(policy Policy, matchesFound bool) {
	statusHandlerMu.RLock()
	handler := policyStatusHandler
	statusHandlerMu.RUnlock()

	handler(policy, matchesFound)
}
