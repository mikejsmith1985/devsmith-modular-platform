package portal_templates

// Small type definitions used by generated templ files. These are minimal
// shims so the generated template code compiles. If richer models are
// required, move these definitions into a shared package and expand fields.

// DashboardUser represents the portal user shown on the dashboard.
type DashboardUser struct {
	Username  string
	Email     string
	AvatarURL string
}

// ServiceInfo describes a service card in the portal dashboard.
type ServiceInfo struct {
	Name        string
	Description string
	URL         string
	Icon        string
	Status      string
}

// LogsDashboardData is the input model for the Logs dashboard template.
type LogsDashboardData struct {
	User DashboardUser
	// Additional fields expected by generated templates
	Stats     map[string]interface{}
	TopErrors []interface{}
	Trends    []interface{}
}
