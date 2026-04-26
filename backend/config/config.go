package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port           string `yaml:"port"`
	DBDriver       string `yaml:"db_driver"`
	DBDSN          string `yaml:"db_dsn"`
	DBTLSMode      string `yaml:"db_tls_mode"`    // disable | require | verify-ca | verify-full
	DBTLSCACert    string `yaml:"db_tls_ca_cert"` // path to CA certificate file
	DBTLSCert      string `yaml:"db_tls_cert"`    // path to client certificate file (mTLS)
	DBTLSKey       string `yaml:"db_tls_key"`     // path to client key file (mTLS)
	TLSCert        string `yaml:"tls_cert"`       // path to server TLS certificate; enables HTTPS when set together with tls_key
	TLSKey         string `yaml:"tls_key"`        // path to server TLS private key
	JWTSecret      string `yaml:"jwt_secret"`
	AllowedOrigins  string `yaml:"allowed_origins"`
	TrustedProxies  string `yaml:"trusted_proxies"` // comma-separated CIDRs/IPs; empty = trust no proxy headers
	WebDir         string `yaml:"web_dir"`
	RedisURL       string `yaml:"redis_url"` // optional — enables horizontal scaling
	DefaultLocale  string `yaml:"default_locale"`
	GinMode        string `yaml:"gin_mode"` // debug | release (default: release)
	DBLog          string `yaml:"db_log"`   // silent | error | warn | info (default: info)
	APILog         bool   `yaml:"api_log"`  // log HTTP requests (default: true)
	UploadDir      string `yaml:"upload_dir"`    // directory for uploaded files (default: ./uploads)
	MaxUploadMB    int64  `yaml:"max_upload_mb"` // max upload size in MB (default: 25)
	BaseURL        string `yaml:"base_url"`      // public base URL (e.g. https://desk.example.com) — used in Swagger UI
	SMTP           SMTPConfig `yaml:"smtp"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	From     string `yaml:"from"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	UseTLS   bool   `yaml:"use_tls"`
}

// Load reads configuration with the following priority (highest first):
//  1. Environment variables
//  2. Config file (path from --config flag, CONFIG_FILE env var, or warmdesk.yaml)
//  3. Built-in defaults
//
// Pass the value of the --config CLI flag as configPath; an empty string
// falls back to the CONFIG_FILE env var and then to "warmdesk.yaml".
func Load(configPath string) *Config {
	cfg := defaults()
	loadFile(cfg, configPath)
	applyEnv(cfg)
	return cfg
}

func defaults() *Config {
	return &Config{
		Port:           "8080",
		DBDriver:       "sqlite",
		DBDSN:          "./warmdesk.db",
		JWTSecret:      "change-me-in-production",
		AllowedOrigins: "http://localhost:5173",
		WebDir:         "",
		GinMode:        "release",
		DefaultLocale:  "en",
		APILog:         true,
		UploadDir:      "./uploads",
		MaxUploadMB:    25,
		SMTP:           SMTPConfig{Port: 587},
	}
}

func loadFile(cfg *Config, flagPath string) {
	// Priority: CLI flag > CONFIG_FILE env var > default filename
	path := flagPath
	if path == "" {
		path = os.Getenv("CONFIG_FILE")
	}
	if path == "" {
		if _, err := os.Stat("warmdesk.yaml"); err == nil {
			path = "warmdesk.yaml"
		} else {
			path = "warmdesk.yml"
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Config file is optional; absence is not an error unless explicitly specified.
		if flagPath != "" {
			log.Fatalf("config: cannot read file %q: %v", path, err)
		}
		return
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Printf("warning: could not parse config file %s: %v", path, err)
	} else {
		log.Printf("loaded config from %s", path)
	}
}

// applyEnv overrides config values with environment variables when set.
func applyEnv(cfg *Config) {
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DB_DRIVER"); v != "" {
		cfg.DBDriver = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.DBDSN = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		cfg.AllowedOrigins = v
	}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		cfg.TrustedProxies = v
	}
	if v := os.Getenv("WEB_DIR"); v != "" {
		cfg.WebDir = v
	}
	if v := os.Getenv("REDIS_URL"); v != "" {
		cfg.RedisURL = v
	}
	if v := os.Getenv("DEFAULT_LOCALE"); v != "" {
		cfg.DefaultLocale = v
	}
	if v := os.Getenv("GIN_MODE"); v != "" {
		cfg.GinMode = v
	}
	if v := os.Getenv("DB_LOG"); v != "" {
		cfg.DBLog = v
	}
	if v := os.Getenv("API_LOG"); v != "" {
		cfg.APILog = v != "false" && v != "0"
	}
	if v := os.Getenv("DB_TLS_MODE"); v != "" {
		cfg.DBTLSMode = v
	}
	if v := os.Getenv("DB_TLS_CA_CERT"); v != "" {
		cfg.DBTLSCACert = v
	}
	if v := os.Getenv("DB_TLS_CERT"); v != "" {
		cfg.DBTLSCert = v
	}
	if v := os.Getenv("DB_TLS_KEY"); v != "" {
		cfg.DBTLSKey = v
	}
	if v := os.Getenv("TLS_CERT"); v != "" {
		cfg.TLSCert = v
	}
	if v := os.Getenv("TLS_KEY"); v != "" {
		cfg.TLSKey = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
}
