package templates

import (
	"context"
	"fmt"
	"net/http"
)

// DashboardUser represents a user displayed on the dashboard.
// It includes the username, email, and avatar URL.
type DashboardUser struct {
	Username  string // The username of the user.
	Email     string // The email address of the user.
	AvatarURL string // The URL of the user's avatar image.
}

// DashboardComponent represents the dashboard UI component.
// It contains the user information to be displayed.
type DashboardComponent struct {
	User DashboardUser // The user information to display on the dashboard.
}

// Dashboard creates a new DashboardComponent for the given user.
// It initializes the component with the provided user data.
func Dashboard(user DashboardUser) *DashboardComponent {
	return &DashboardComponent{User: user}
}

// Render generates the HTML for the dashboard and writes it to the response writer.
// It uses the user information stored in the DashboardComponent.
// If an error occurs while writing the response, it logs the error and sends an HTTP 500 status.
func (d *DashboardComponent) Render(ctx context.Context, w http.ResponseWriter) {
	// Debug logging
	fmt.Printf("Rendering Dashboard for user: %+v\n", d.User)

	// Render the dashboard HTML template (simplified for now)
	html := `<!DOCTYPE html>
<html>
<head><title>Dashboard</title></head>
<body>
	<h1>Welcome, ` + d.User.Username + `</h1>
	<p>Email: ` + d.User.Email + `</p>
	<img src="` + d.User.AvatarURL + `" alt="Avatar">
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		// Improved error handling
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		fmt.Printf("[ERROR] Failed to write response: %v\n", err)
	}
}
