// Package tokens defines the design system tokens for the DevSmith platform.
// Tokens include colors, spacing, typography, shadows, and transitions.
package tokens

// ColorPalette represents the color system with light and dark mode support.
type ColorPalette struct {
	// Primary colors
	Primary       string // #007AFF - Primary action color
	PrimaryHover  string // #0051D5 - Primary hover state
	PrimaryActive string // #003DA3 - Primary active state

	// Semantic colors
	Success string // #34C759 - Success state
	Warning string // #FF9500 - Warning state
	Danger  string // #FF3B30 - Danger/Error state
	Info    string // #00C7E7 - Information state

	// Neutral colors (light mode)
	Background    string // #FFFFFF - Main background
	Surface       string // #F5F5F7 - Secondary background
	SurfaceSecond string // #ECECF1 - Tertiary background
	Border        string // #D2D2D7 - Border color
	Text          string // #1D1D1F - Primary text
	TextSecondary string // #86868B - Secondary text
	TextTertiary  string // #A1A1A6 - Tertiary text

	// Dark mode overrides
	DarkBackground string // #000000
	DarkSurface    string // #1C1C1E
	DarkSurface2   string // #2C2C2E
	DarkBorder     string // #424245
	DarkText       string // #F5F5F7
	DarkText2      string // #98989D
	DarkText3      string // #8E8E93
}

// Spacing scale - 8px base unit
type Spacing struct {
	// Base units (multiples of 2px for fine control)
	XS    string // 2px
	Small string // 4px

	// Standard units (multiples of 8px)
	Base   string // 8px (1 unit)
	Half   string // 4px (0.5 unit)
	Double string // 16px (2 units)
	Triple string // 24px (3 units)
	Quad   string // 32px (4 units)
	Five   string // 40px (5 units)
	Six    string // 48px (6 units)

	// Large spacing
	Seven string // 56px
	Eight string // 64px
	Ten   string // 80px
}

// Typography tokens
type Typography struct {
	// Font families
	SystemFont string // -apple-system, BlinkMacSystemFont, Segoe UI, Helvetica, Arial
	MonoFont   string // Menlo, Monaco, Courier New

	// Font sizes (rem-based for accessibility)
	Size12 string // 0.75rem (12px)
	Size13 string // 0.8125rem (13px)
	Size14 string // 0.875rem (14px)
	Size15 string // 0.9375rem (15px)
	Size16 string // 1rem (16px)
	Size18 string // 1.125rem (18px)
	Size20 string // 1.25rem (20px)
	Size24 string // 1.5rem (24px)
	Size28 string // 1.75rem (28px)
	Size32 string // 2rem (32px)
	Size36 string // 2.25rem (36px)

	// Font weights
	Regular  string // 400
	Medium   string // 500
	SemiBold string // 600
	Bold     string // 700
	Heavy    string // 800

	// Line heights
	Tight   string // 1.2
	Normal  string // 1.5
	Relaxed string // 1.75
	Loose   string // 2
}

// BorderRadius represents border radius token values.
type BorderRadius struct {
	None   string // 0px
	Small  string // 4px
	Medium string // 6px
	Large  string // 8px
	XL     string // 12px
	Full   string // 9999px (full circle)
}

// Shadow tokens
type Shadow struct {
	// Light mode shadows
	Shallow string // 0 1px 3px rgba(0,0,0,0.1)
	Small   string // 0 2px 4px rgba(0,0,0,0.1)
	Medium  string // 0 4px 12px rgba(0,0,0,0.15)
	Large   string // 0 8px 24px rgba(0,0,0,0.2)
	XL      string // 0 16px 40px rgba(0,0,0,0.25)

	// Dark mode shadows
	DarkShallow string // 0 1px 3px rgba(0,0,0,0.3)
	DarkSmall   string // 0 2px 4px rgba(0,0,0,0.4)
	DarkMedium  string // 0 4px 12px rgba(0,0,0,0.5)
	DarkLarge   string // 0 8px 24px rgba(0,0,0,0.6)
}

// Transition tokens for animations
type Transition struct {
	Fast   string // 100ms cubic-bezier(0.4, 0, 0.2, 1)
	Base   string // 200ms cubic-bezier(0.4, 0, 0.2, 1)
	Slow   string // 300ms cubic-bezier(0.4, 0, 0.2, 1)
	Slower string // 500ms cubic-bezier(0.4, 0, 0.2, 1)
}

// Tokens represents the complete design token system - the single source of truth.
type Tokens struct {
	Colors       *ColorPalette
	Spacing      *Spacing
	Typography   *Typography
	BorderRadius *BorderRadius
	Shadows      *Shadow
	Transitions  *Transition
}

// NewTokens creates a new design token system
func NewTokens() *Tokens {
	return &Tokens{
		Colors: &ColorPalette{
			Primary:        "#007AFF",
			PrimaryHover:   "#0051D5",
			PrimaryActive:  "#003DA3",
			Success:        "#34C759",
			Warning:        "#FF9500",
			Danger:         "#FF3B30",
			Info:           "#00C7E7",
			Background:     "#FFFFFF",
			Surface:        "#F5F5F7",
			SurfaceSecond:  "#ECECF1",
			Border:         "#D2D2D7",
			Text:           "#1D1D1F",
			TextSecondary:  "#86868B",
			TextTertiary:   "#A1A1A6",
			DarkBackground: "#000000",
			DarkSurface:    "#1C1C1E",
			DarkSurface2:   "#2C2C2E",
			DarkBorder:     "#424245",
			DarkText:       "#F5F5F7",
			DarkText2:      "#98989D",
			DarkText3:      "#8E8E93",
		},
		Spacing: &Spacing{
			XS:     "2px",
			Small:  "4px",
			Base:   "8px",
			Half:   "4px",
			Double: "16px",
			Triple: "24px",
			Quad:   "32px",
			Five:   "40px",
			Six:    "48px",
			Seven:  "56px",
			Eight:  "64px",
			Ten:    "80px",
		},
		Typography: &Typography{
			SystemFont: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
			MonoFont:   "'Menlo', 'Monaco', 'Courier New', monospace",
			Size12:     "0.75rem",
			Size13:     "0.8125rem",
			Size14:     "0.875rem",
			Size15:     "0.9375rem",
			Size16:     "1rem",
			Size18:     "1.125rem",
			Size20:     "1.25rem",
			Size24:     "1.5rem",
			Size28:     "1.75rem",
			Size32:     "2rem",
			Size36:     "2.25rem",
			Regular:    "400",
			Medium:     "500",
			SemiBold:   "600",
			Bold:       "700",
			Heavy:      "800",
			Tight:      "1.2",
			Normal:     "1.5",
			Relaxed:    "1.75",
			Loose:      "2",
		},
		BorderRadius: &BorderRadius{
			None:   "0px",
			Small:  "4px",
			Medium: "6px",
			Large:  "8px",
			XL:     "12px",
			Full:   "9999px",
		},
		Shadows: &Shadow{
			Shallow:     "0 1px 3px rgba(0,0,0,0.1)",
			Small:       "0 2px 4px rgba(0,0,0,0.1)",
			Medium:      "0 4px 12px rgba(0,0,0,0.15)",
			Large:       "0 8px 24px rgba(0,0,0,0.2)",
			XL:          "0 16px 40px rgba(0,0,0,0.25)",
			DarkShallow: "0 1px 3px rgba(0,0,0,0.3)",
			DarkSmall:   "0 2px 4px rgba(0,0,0,0.4)",
			DarkMedium:  "0 4px 12px rgba(0,0,0,0.5)",
			DarkLarge:   "0 8px 24px rgba(0,0,0,0.6)",
		},
		Transitions: &Transition{
			Fast:   "100ms cubic-bezier(0.4, 0, 0.2, 1)",
			Base:   "200ms cubic-bezier(0.4, 0, 0.2, 1)",
			Slow:   "300ms cubic-bezier(0.4, 0, 0.2, 1)",
			Slower: "500ms cubic-bezier(0.4, 0, 0.2, 1)",
		},
	}
}
