package config

import (
	"fmt"
	"os"
	"encoding/json"
)

const (
	AppVersion = "2020.1.0"

	AccessAllow = 1
	AccessDeny = 0
)

type AccessLevel int
func (l *AccessLevel) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str); if err != nil {
		return err
	}
	lvl, err := stringToAccessLevel(str); if err != nil {
		return err
	}

	*l = lvl
	return nil
}

type AppConfig struct {
	Address string
	Port int
	Logfile string
	Loglevel string
	CorsOrigin string `json:"cors-origin"`
	AccessPolicy AccessPolicy `json:"access-policy"`
	Database Database
}

type Database struct {
	Protocol string
	Host string
	Port int
	User string
	Password string
	DatabaseName string `json:"database"`
}

type AccessPolicy struct {
	Default AccessLevel
	Rules []AccessRule
}

type AccessRule struct {
	Origin string
	Policy AccessLevel
}

var instance *AppConfig

func GetAppConfig() *AppConfig {
	if instance == nil {
		instance = &AppConfig{
			Address: "0.0.0.0",
			Port: 7534,
			Logfile: "./shoptrac.log",
			Loglevel: "debug",
			CorsOrigin: "*",
			AccessPolicy: AccessPolicy{
				Default: AccessAllow,
				Rules: nil,
			},
			Database: Database{
				Host: "localhost",
				Port: 8529,
				User: "dummy",
				Password: "doof",
				DatabaseName: "shoptrac",
			},
		}
	}

	return instance
}

func (c *AppConfig) LoadFromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}

	f, err := os.Open(filename); if err != nil {
		return fmt.Errorf("failed to open file '%s': %s", filename, err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(c); if err != nil {
		return fmt.Errorf("failed to decode file: %s", err)
	}

	return nil
}

func (c *AppConfig) String() (string, error) {
	res, err := json.Marshal(c); if err != nil {
		return "", fmt.Errorf("failed to encode AppConfig to JSON: %s", err)
	}
	return string(res), nil
}

func stringToAccessLevel(level string) (AccessLevel, error) {
	switch level {
	case "allow":
		return AccessAllow, nil

	case "deny":
		return AccessDeny, nil
		
	default:
		return -1, fmt.Errorf("unknown access level '%s'", level)
	}
}