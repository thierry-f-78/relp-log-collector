package config

import "fmt"
import "os"
import "time"

import "gopkg.in/yaml.v2"

type ACL struct {
	Value  string `yaml:"value"`
	Action string `yaml:"action"`
}

type RELP struct {
	Certificate string `yaml:"certificate"`
	PrivateKey  string `yaml:"private-key"`
	CA          string `yaml:"ca"`
	Listen      string `yaml:"listen"`
	ACL         []ACL  `yaml:"acl"`
}

type Clickhouse struct {
	Target []string `yaml:"target"`
}

type PostgreSQL struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
	Timeout  uint   `yaml:"timeout"`
}

type SQLite struct {
	Path string `yaml:"path"`
}

type Spool struct {
	Path    string        `yaml:"path"`
	MaxLogs uint          `yaml:"max-logs"`
	MaxIdle time.Duration `yaml:"max-idle"`
}

type Dispatch struct {
	MinLogs       uint          `yaml:"min-logs"`
	MaxWait       time.Duration `yaml:"max-wait"`
	CheckInterval time.Duration `yaml:"check-interval"`
}

type Plugins struct {
	Path []string `yaml:"path"`
}

type Config struct {
	RELP       RELP        `yaml:"relp"`
	Clickhouse *Clickhouse `yaml:"clickhouse"`
	PostgreSQL *PostgreSQL `yaml:"postgresql"`
	SQLite     *SQLite     `yaml:"sqlite"`
	Spool      Spool       `yaml:"spool"`
	Dispatch   Dispatch    `yaml:"dispatch"`
	Plugins    Plugins     `yaml:"plugins"`
}

var Cf Config

func Load(configFile string) error {
	var data []byte
	var err error

	// Check configuraiton file is not empty
	if configFile == "" {
		return fmt.Errorf("No configuration file specified")
	}

	// Read yaml file
	data, err = os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("Error loading configuration file %q: %v", configFile, err)
	}

	// Decode yaml
	err = yaml.Unmarshal(data, &Cf)
	if err != nil {
		return fmt.Errorf("Error decoding config file %q YAML: %v", configFile, err)
	}

	return nil
}
