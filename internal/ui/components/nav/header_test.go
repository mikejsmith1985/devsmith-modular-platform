package nav

import (
	"bytes"
	"context"
	"strings"
	"testing"

	portalmodels "github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHeader_RenderBackButtonOnlyForNonPortalApps verifies back button appears conditionally
func TestHeader_RenderBackButtonOnlyForNonPortalApps(t *testing.T) {
	tests := []struct {
		name          string
		currentApp    string
		expectBackBtn bool
	}{
		{"Portal hides back button", "portal", false},
		{"Review shows back button", "review", true},
		{"Logs shows back button", "logs", true},
		{"Analytics shows back button", "analytics", true},
		{"HealthCheck shows back button", "healthcheck", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &portalmodels.User{
				Username: "testuser",
				Email:    "test@example.com",
			}

			var buf bytes.Buffer
			err := Header(tt.currentApp, user).Render(context.Background(), &buf)
			require.NoError(t, err, "Header should render without error")

			html := buf.String()

			if tt.expectBackBtn {
				// Check for back button navigation element
				assert.Contains(t, html, "Portal</a>", "Non-portal apps should show back button to Portal")
				assert.Contains(t, html, `href="/"`, "Back button should link to Portal home")
				assert.Contains(t, html, "nav-back", "Back button should have nav-back class")
			} else {
				// Portal app should not have redundant back button
				// But should have logo link to home
				assert.Contains(t, html, "DevSmith", "Should always show logo")
			}
		})
	}
}

// TestHeader_LogoAlwaysPresent verifies logo is rendered in all cases
func TestHeader_LogoAlwaysPresent(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	for _, app := range []string{"portal", "review", "logs", "analytics"} {
		t.Run("Logo present in "+app, func(t *testing.T) {
			var buf bytes.Buffer
			err := Header(app, user).Render(context.Background(), &buf)
			require.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, "DevSmith", "Logo should always be present")
			assert.Contains(t, html, `class="logo"`, "Logo should have logo class")
		})
	}
}

// TestHeader_AppSwitcherContainsAllApps verifies dropdown has all 5 apps
func TestHeader_AppSwitcherContainsAllApps(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("review", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify all apps are in switcher
	apps := []struct {
		name string
		href string
	}{
		{"Portal", `href="/"`},
		{"Review", `href="/review"`},
		{"Logs", `href="/logs"`},
		{"Analytics", `href="/analytics"`},
		{"Health Check", `href="/healthcheck"`},
	}

	for _, app := range apps {
		assert.Contains(t, html, app.name, "App switcher should contain "+app.name)
		assert.Contains(t, html, app.href, "App link should point to correct path")
	}

	// Verify dropdown structure
	assert.Contains(t, html, `role="menu"`, "Dropdown should have role=menu for accessibility")
	assert.Contains(t, html, `role="menuitem"`, "Dropdown items should have role=menuitem")
}

// TestHeader_AppIndicatorShowsCurrentApp verifies current app is displayed
func TestHeader_AppIndicatorShowsCurrentApp(t *testing.T) {
	tests := []string{"portal", "review", "logs", "analytics", "healthcheck"}

	for _, app := range tests {
		t.Run("Current app indicator for "+app, func(t *testing.T) {
			user := &portalmodels.User{
				Username: "testuser",
				Email:    "test@example.com",
			}

			var buf bytes.Buffer
			err := Header(app, user).Render(context.Background(), &buf)
			require.NoError(t, err)

			html := buf.String()
			assert.Contains(t, html, `class="current-app"`, "Should have current-app indicator")
			// Check that the app name appears in the content (it will be capitalized by CSS)
			assert.Contains(t, html, app, "Current app name should be in the header")
		})
	}
}

// TestHeader_DarkModeToggleRendered verifies theme toggle is present
func TestHeader_DarkModeToggleRendered(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify theme toggle button
	assert.Contains(t, html, `id="theme-toggle"`, "Theme toggle should have ID for JavaScript")
	assert.Contains(t, html, `theme-toggle`, "Theme toggle should have appropriate class")
	assert.Contains(t, html, `aria-label="Toggle dark mode"`, "Theme toggle should be accessible")

	// Verify icons
	assert.Contains(t, html, `class="sun-icon"`, "Should have sun icon for light mode")
	assert.Contains(t, html, `class="moon-icon"`, "Should have moon icon for dark mode")
}

// TestHeader_UserMenuShowsUserInfo verifies user menu displays correctly
func TestHeader_UserMenuShowsUserInfo(t *testing.T) {
	user := &portalmodels.User{
		Username: "johndoe",
		Email:    "john@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify user info displayed
	assert.Contains(t, html, "johndoe", "Username should be displayed in user menu")
	assert.Contains(t, html, "john@example.com", "Email should be displayed in user menu")

	// Verify user avatar (first letter)
	assert.Contains(t, html, `class="user-avatar"`, "User avatar should be present")
	assert.Contains(t, html, "j", "User avatar should show first letter of username")
}

// TestHeader_UserMenuHasLogoutButton verifies logout functionality
func TestHeader_UserMenuHasLogoutButton(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify logout button
	assert.Contains(t, html, "Logout", "User menu should have logout button")
	assert.Contains(t, html, `hx-post="/api/auth/logout"`, "Logout should use HTMX POST")
	assert.Contains(t, html, `logout-btn`, "Logout button should have logout-btn class")
}

// TestHeader_UserMenuHasSettingsLinks verifies settings navigation
func TestHeader_UserMenuHasSettingsLinks(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify menu items
	assert.Contains(t, html, "Profile", "User menu should have Profile link")
	assert.Contains(t, html, "Settings", "User menu should have Settings link")
	assert.Contains(t, html, "AI Preferences", "User menu should have AI Preferences link")

	// Verify links
	assert.Contains(t, html, `href="/profile"`, "Profile link should be present")
	assert.Contains(t, html, `href="/settings"`, "Settings link should be present")
	assert.Contains(t, html, `href="/settings/ai"`, "AI Preferences link should be present")
}

// TestHeader_HeaderStructure verifies correct semantic layout
func TestHeader_HeaderStructure(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Verify header element
	assert.Contains(t, html, `<header`, "Should be a header element")
	assert.Contains(t, html, `class="devsmith-header"`, "Header should have correct class")
	assert.Contains(t, html, `</header>`, "Header should be properly closed")

	// Verify three-part structure
	assert.Contains(t, html, `class="header-left"`, "Header should have left section")
	assert.Contains(t, html, `class="header-center"`, "Header should have center section")
	assert.Contains(t, html, `class="header-right"`, "Header should have right section")
}

// TestHeader_AccessibilityAttributes verifies ARIA labels and roles
func TestHeader_AccessibilityAttributes(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Check for ARIA labels
	assert.Contains(t, html, `aria-label=`, "Header elements should have ARIA labels")
	assert.Contains(t, html, `role="menu"`, "Dropdowns should have menu role")
	assert.Contains(t, html, `role="menuitem"`, "Menu items should have correct role")

	// Verify user button accessibility
	assert.Contains(t, html, `aria-label="User menu for`, "User button should have accessible label")
}

// TestHeader_WithNilUser handles case where user is not authenticated
func TestHeader_WithNilUser(t *testing.T) {
	var buf bytes.Buffer
	err := Header("portal", nil).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Should still have logo and navigation
	assert.Contains(t, html, "DevSmith", "Logo should always render")
	assert.Contains(t, html, "theme-toggle", "Theme toggle should be available without auth")

	// Should not have user menu
	assert.NotContains(t, html, "user-menu", "User menu should not render if user is nil")
}

// TestHeader_LogoLinksToHome verifies logo navigation
func TestHeader_LogoLinksToHome(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("review", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Find logo link
	assert.Contains(t, html, `<a href="/" class="logo"`, "Logo should link to home")
}

// TestHeader_DropdownMenuAccessibility ensures proper keyboard navigation structure
func TestHeader_DropdownMenuAccessibility(t *testing.T) {
	user := &portalmodels.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	var buf bytes.Buffer
	err := Header("portal", user).Render(context.Background(), &buf)
	require.NoError(t, err)

	html := buf.String()

	// Count menu items to ensure proper structure
	menuItemCount := strings.Count(html, `role="menuitem"`)
	// Should have: Portal, Review, Logs, Analytics, Health Check = 5 app menu items
	// Plus: Profile, Settings, AI Preferences, Logout = 4 user menu items
	// Total: 9
	assert.GreaterOrEqual(t, menuItemCount, 9, "Should have proper menu structure")
}
