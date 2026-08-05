package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for both export and import operations.
type Config struct {
	WarmDesk WarmDeskConfig `yaml:"warmdesk"`
	Platform PlatformConfig `yaml:"platform"`
	// ColumnMap maps WarmDesk column names → external platform column/status names.
	// If empty or a name is not found, the WarmDesk column name is used as-is.
	ColumnMap map[string]string `yaml:"column_map"`
	// UserMap maps external platform user names (as they appear on cards —
	// e.g. a Ryver display name like "Ton Kersten") to a WarmDesk username
	// (e.g. "tonk"). Used on import to assign cards to the right WarmDesk
	// user. Names not present here are left unassigned.
	UserMap map[string]string `yaml:"user_map"`
}

// WarmDeskConfig holds connection details for the WarmDesk server.
type WarmDeskConfig struct {
	URL       string `yaml:"url"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Project   string `yaml:"project"`    // project slug
	Customer  string `yaml:"customer"`   // exact name of the WarmDesk customer the project belongs to (required)
	KeyPrefix string `yaml:"key_prefix"` // override the auto-derived card prefix (e.g. when two source project names would collide); optional
}

// PlatformConfig holds connection details for the external platform.
type PlatformConfig struct {
	Name string `yaml:"name"` // jira | trello | openproject | ryver

	// Jira
	URL        string `yaml:"url"`
	Email      string `yaml:"email"`
	APIToken   string `yaml:"api_token"`
	ProjectKey string `yaml:"project_key"`
	IssueType  string `yaml:"issue_type"` // default: Task

	// Trello
	APIKey  string `yaml:"api_key"`
	Token   string `yaml:"token"`
	BoardID string `yaml:"board_id"`

	// OpenProject (reuses URL, APIKey)
	ProjectID string `yaml:"project_id"`

	// Ryver
	Org string `yaml:"org"`
	// api_token reuses APIToken above
	Team string `yaml:"team"`
}

// LoadConfig reads the YAML config file and applies environment variable
// overrides. Missing required fields are prompted interactively.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	// Load YAML file if it exists.
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	// Environment variable overrides.
	applyEnv(&cfg.WarmDesk.URL, "WARMDESK_URL")
	applyEnv(&cfg.WarmDesk.Username, "WARMDESK_USERNAME")
	applyEnv(&cfg.WarmDesk.Password, "WARMDESK_PASSWORD")
	applyEnv(&cfg.WarmDesk.Project, "WARMDESK_PROJECT")
	applyEnv(&cfg.WarmDesk.Customer, "WARMDESK_CUSTOMER")
	applyEnv(&cfg.Platform.APIToken, "PLATFORM_API_TOKEN")
	applyEnv(&cfg.Platform.APIKey, "PLATFORM_API_KEY")

	// Interactive prompts for required fields.
	promptIfEmpty(&cfg.WarmDesk.URL, "WarmDesk URL")
	promptIfEmpty(&cfg.WarmDesk.Username, "WarmDesk username")
	promptIfEmpty(&cfg.WarmDesk.Password, "WarmDesk password")
	promptIfEmpty(&cfg.WarmDesk.Project, "WarmDesk project slug")
	promptIfEmpty(&cfg.WarmDesk.Customer, "WarmDesk customer name")

	if cfg.Platform.IssueType == "" {
		cfg.Platform.IssueType = "Task"
	}

	// Trim trailing slashes so URL-joining call sites (baseURL + "/api/v1/...")
	// never produce a double slash — which some reverse proxies route to a
	// fallback/HTML response instead of the Go backend.
	cfg.WarmDesk.URL = strings.TrimRight(cfg.WarmDesk.URL, "/")
	cfg.Platform.URL = strings.TrimRight(cfg.Platform.URL, "/")

	return cfg, nil
}

// ReverseColumnMap returns a map from external platform names → WarmDesk column
// names (the inverse of cfg.ColumnMap).
func ReverseColumnMap(m map[string]string) map[string]string {
	rev := make(map[string]string, len(m))
	for k, v := range m {
		rev[v] = k
	}
	return rev
}

// MapColumn translates a WarmDesk column name to the external name using the
// column map. If no mapping exists the original name is returned unchanged.
func MapColumn(name string, columnMap map[string]string) string {
	if columnMap == nil {
		return name
	}
	if mapped, ok := columnMap[name]; ok {
		return mapped
	}
	return name
}

// MapColumnReverse translates an external column/status name back to a
// WarmDesk column name using the reversed column map.
func MapColumnReverse(name string, reverseMap map[string]string) string {
	if reverseMap == nil {
		return name
	}
	if mapped, ok := reverseMap[name]; ok {
		return mapped
	}
	return name
}

// ─── internal helpers ────────────────────────────────────────────────────────

func applyEnv(field *string, key string) {
	if v := os.Getenv(key); v != "" {
		*field = v
	}
}

var stdinReader = bufio.NewReader(os.Stdin)

func promptIfEmpty(field *string, label string) {
	if *field != "" {
		return
	}
	fmt.Printf("%s: ", label)
	val, _ := stdinReader.ReadString('\n')
	*field = strings.TrimSpace(val)
}
