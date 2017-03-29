package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kardianos/osext"
	"github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
)

// The Config struct holds the Eremetic Configuration
type Config struct {
	// Logging
	LogLevel  string `yaml:"loglevel"`
	LogFormat string `yaml:"logformat"`

	// Server
	Address         string `yaml:"address"`
	Port            int    `yaml:"port"`
	HTTPCredentials string `yaml:"http_credentials" envconfig:"http_credentials"`

	// Database
	DatabaseDriver string `yaml:"database_driver" envconfig:"database_driver"`
	DatabasePath   string `yaml:"database" envconfig:"database"`

	// Mesos
	Name             string  `yaml:"name"`
	User             string  `yaml:"user"`
	Checkpoint       bool    `yaml:"checkpoint"`
	FailoverTimeout  float64 `yaml:"failover_timeout" envconfig:"failover_timeout"`
	QueueSize        int     `yaml:"queue_size" envconfig:"queue_size"`
	Master           string  `yaml:"master"`
	FrameworkID      string  `yaml:"framework_id" envconfig:"framework_id"`
	CredentialsFile  string  `yaml:"credential_file" envconfig:"credential_file"`
	MessengerAddress string  `yaml:"messenger_address" envconfig:"messenger_address"`
	MessengerPort    int     `yaml:"messenger_port" envconfig:"messenger_port"`
}

// DefaultConfig returns a Config struct with the default settings
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  "debug",
		LogFormat: "text",

		DatabaseDriver: "boltdb",
		DatabasePath:   "db/eremetic.db",

		Name:            "Eremetic",
		User:            "root",
		Checkpoint:      true,
		FailoverTimeout: 2592000.0,
		QueueSize:       100,
	}
}

// GetConfigFilePath returns the location of the config file in order of priority:
// 1 ) File in same directory as the executable
// 2 ) Global file in /etc/eremetic/eremetic.yml
func GetConfigFilePath() string {
	path, _ := osext.ExecutableFolder()
	path = fmt.Sprintf("%s/eremetic.yml", path)
	if _, err := os.Open(path); err == nil {
		return path
	}
	globalPath := "/etc/eremetic/eremetic.yml"
	if _, err := os.Open(globalPath); err == nil {
		return globalPath
	}

	return ""
}

// ReadConfigFile reads the config file and overrides any values net in both it
// and the DefaultConfig
func ReadConfigFile(conf *Config, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	configFile, _ := ioutil.ReadAll(file)
	yaml.Unmarshal(configFile, conf)

	if conf.DatabaseDriver == "boltdb" && conf.DatabasePath == "" {
		conf.DatabasePath = "db/eremetic.db"
	}
}

// ReadEnvironment takes environment variables and overrides any values from
// DefaultConfig and the Config file.
func ReadEnvironment(conf *Config) {
	envconfig.Process("", conf)
}
