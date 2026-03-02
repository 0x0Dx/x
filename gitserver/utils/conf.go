package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	App      AppConfig      `json:"app"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

type AppConfig struct {
	Name string `json:"name"`
}

type ServerConfig struct {
	HTTPAddr string `json:"http_addr"`
	HTTPPort string `json:"http_port"`
}

type DatabaseConfig struct {
	DBType string `json:"db_type"`
	Host   string `json:"host"`
	Name   string `json:"name"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

var Cfg *Config

func exeDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

func init() {
	workDir, err := exeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fail to get work directory: %v\n", err)
		os.Exit(2)
	}

	cfgPath := filepath.Join(workDir, "conf", "app.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		cfgPath = filepath.Join(workDir, "..", "conf", "app.json")
		data, err = os.ReadFile(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot load config file: %v\n", err)
			os.Exit(2)
		}
	}

	Cfg = new(Config)
	if err := json.Unmarshal(data, Cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse config file: %v\n", err)
		os.Exit(2)
	}
}
