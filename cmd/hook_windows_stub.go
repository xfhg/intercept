//go:build windows && amd64
// +build windows,amd64

package cmd

// Windows amd64 stubs for webhook functions.
// These are no-ops so the Windows build compiles and runs without attempting
// webhook delivery. Change to a real implementation if you want webhooks on Windows.

// GenerateWebhookSecret is a noop on Windows builds. Return an empty secret and no error.
func GenerateWebhookSecret() (string, error) {
	return "", nil
}

// PostReportToWebhooks is a noop on Windows builds.
func PostReportToWebhooks(sarifReport SARIFReport) error {
	return nil
}

// PostResultsToWebhooks is a noop on Windows builds.
// This function is referenced by runtime/observe flows; provide it to avoid
// undefined symbol errors on Windows amd64.
func PostResultsToWebhooks(sarifReport SARIFReport) error {
	return nil
}

// Optional explicit error variant (uncomment to make webhook calls fail loudly):
// func PostResultsToWebhooks(sarifReport SARIFReport) error {
//     return fmt.Errorf("webhook posting not supported on Windows amd64 builds")
// }
