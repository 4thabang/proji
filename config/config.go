package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// APIAuthentication represents the configurable and authentication related values in the main config.
type APIAuthentication struct {
	GHToken string `mapstructure:"gh_token"`
	GLToken string `mapstructure:"gl_token"`
}

// DatabaseConnection represents the configurable and database related values in the main config.
type DatabaseConnection struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// Config represents central resources and information the app uses.
type Config struct {
	Auth               *APIAuthentication  `mapstructure:"auth"`
	BasePath           string              `mapstructure:"-"`
	DatabaseConnection *DatabaseConnection `mapstructure:"database"`
	ExcludedPaths      []string            `mapstructure:"import.exclude_folders"`
	provider           *viper.Viper        `mapstructure:"-"`
}

const (
	defaultDatabaseDriver = "sqlite3"
	defaultDatabaseDSN    = "/db/proji.sqlite3"
)

//nolint:gochecknoglobals
var globalBasePath string

// Setup determines the operating system specific base config path and stores it. This needs to be run before all other
// config methods.
func Setup() error {
	// Load and set the config base path
	return setGlobalBasePath()
}

// New returns a new empty config instance which has its base path set to the given path.
func New(path string) *Config {
	// Set platform specific config path
	conf := &Config{}
	conf.BasePath = path
	return conf
}

// Load tries to load all configuration values.
func (c *Config) Load() error {
	// Set config provider
	c.setProvider()

	// Set config specifications
	c.setSpecs()

	// Set default config values
	c.setDefaultValues()

	// Load config values
	err := c.loadValues()
	if err != nil {
		return err
	}

	// Set the loaded values as config
	err = c.setFinalValues()
	if err != nil {
		return err
	}

	// Handle special case for sqlite3 database
	c.handleDatabaseDriverSpecialCase()
	return nil
}

// setProvider sets the configs provider to a new viper instance.
func (c *Config) setProvider() {
	c.provider = viper.New()
}

// setSpecs sets config provider specifications like the config path and env prefix.
func (c *Config) setSpecs() {
	c.provider.AddConfigPath(c.BasePath)
	c.provider.SetConfigName("config")
	c.provider.SetConfigType("toml")
	c.provider.SetEnvPrefix("PROJI")
	c.provider.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// setDefaultValues sets the main configs default values for all keys.
func (c *Config) setDefaultValues() {
	c.provider.SetDefault("auth.gh_token", "")
	c.provider.SetDefault("auth.gl_token", "")
	c.provider.SetDefault("import.exclude_folders", []string{})
	c.provider.SetDefault("database.driver", defaultDatabaseDriver)
	c.provider.SetDefault("database.dsn", filepath.Join(c.BasePath, defaultDatabaseDSN))
}

// set should run after loadFile and loadEnvironmentVariables. It sets the loaded values as the final config.
func (c *Config) setFinalValues() error {
	return c.provider.Unmarshal(c)
}

// loadValues loads config values from file and environment variables.
func (c *Config) loadValues() error {
	err := c.loadFile()
	if err != nil {
		return err
	}
	c.loadEnvironmentVariables()
	return nil
}

// loadConfigFile tries to load the settings relevant values from the default config file. Skips if config file not found.
func (c *Config) loadFile() error {
	err := c.provider.ReadInConfig()
	_, ok := err.(viper.ConfigFileNotFoundError)
	if ok {
		// Config file not found; ignore error and return empty settings
		return nil
	}
	return err
}

// loadSettingsFromConfig tries to load the settings from the default config file. Skips if config file not found.
func (c *Config) loadEnvironmentVariables() {
	c.provider.AutomaticEnv()
}

func (c *Config) handleDatabaseDriverSpecialCase() {
	// Special case for sqlite.
	if c.DatabaseConnection.Driver == "sqlite3" {
		c.DatabaseConnection.DSN = RelativePathToAbsoluteConfigPath(c.BasePath, c.DatabaseConnection.DSN)
	}
}

// GetBaseConfigPath returns the OS specific base path of the config folder.
func GetBaseConfigPath() string {
	return globalBasePath
}

// setGlobalBasePath sets the variable globalBasePath to the OS specific base path of the config folder.
func setGlobalBasePath() error {
	if globalBasePath != "" {
		return nil
	}
	var path string
	var err error
	switch runtime.GOOS {
	case "linux":
		path, err = getLinuxConfigBasePath()
	case "darwin":
		path, err = getDarwinConfigBasePath()
	case "windows":
		path, err = getWindowsConfigBasePath()
	default:
		err = fmt.Errorf("OS %s is not supported and/or tested yet. Please create an issue at "+
			"https://github.com/nikoksr/proji to request the support of your OS", runtime.GOOS)
	}
	if err != nil {
		return err
	}
	// No errors, set the global base path
	globalBasePath = path
	return nil
}

// getLinuxConfigBasePath tries to read the HOME env variable. Returns proji's home path on linux systems on success.
func getLinuxConfigBasePath() (string, error) {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		return "", fmt.Errorf("could not find environment variable HOME")
	}
	return filepath.Join(home, "/.config/proji"), nil
}

// getDarwinConfigBasePath tries to read the HOME env variable. Returns proji's home path on darwin systems on success.
func getDarwinConfigBasePath() (string, error) {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		return "", fmt.Errorf("could not find environment variable HOME")
	}
	return filepath.Join(home, "/Library/Application Support/proji"), nil
}

// getWindowsConfigBasePath tries to read the APPDATA env variable. Returns proji's home path on windows systems on success.
func getWindowsConfigBasePath() (string, error) {
	appData, exists := os.LookupEnv("APPDATA")
	if !exists {
		return "", fmt.Errorf("could not find environment variable APPDATA")
	}
	return filepath.Join(appData, "/proji"), nil
}

// RelativePathToAbsoluteConfigPath takes a relative path and a config folder path (usually proji's main config folder)
// and returns the absolute path of the relative path in relation to the config folder path. Returns the relative path
// unchanged if it was already an absolute path.
func RelativePathToAbsoluteConfigPath(configFolderPath, relativePath string) string {
	if filepath.IsAbs(relativePath) {
		// Either user defined path like '/my/custom/db/path' or default value was loaded
		return relativePath
	}
	// User defined path like 'db/proji.sqlite3'. Gets prefixed with config folder path. This has to be a relative
	// path or else the above will trigger.
	return filepath.Join(configFolderPath, relativePath)
}
