package gronx

// Minimal shim of the github.com/adhocore/gronx root package used by intercept.
// The real package exposes a New() function that returns a struct with IsValid(expr string) bool.
// This shim implements a tiny compatible subset so we can build locally and avoid
// the upstream syscall/Setpgid issues on Windows when the full module causes build errors.

// Cron provides IsValid for cron expressions. Keep it extremely small and permissive.
type Cron struct{}

// IsValid returns true when expr is non-empty. This is intentionally minimal.
// If you need stricter validation on your CI, replace this with a cron parser.
func (c *Cron) IsValid(expr string) bool {
	return expr != ""
}

// New returns a new Cron validator.
func New() *Cron {
	return &Cron{}
}
