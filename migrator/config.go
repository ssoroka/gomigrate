package migrator

import (
	"encoding/json"
	"go/build"
	"os"
)

var ConfigFileName = ".migrate"

type Config struct {
	ProjectPath            string
	GoMigratePackagePath   string
	MainMigrationFile      string
	LocalMigratorPath      string
	LocalMigrationsPath    string
	LocalMigrationsPackage string
	LocalTemplatesPath     string
}

func DefaultConfig() *Config {
	config := &Config{
		GoMigratePackagePath: os.Getenv("GOPATH"),
	}
	if config.GoMigratePackagePath == "" {
		config.GoMigratePackagePath = build.Default.GOPATH + "/src/github.com/ssoroka/gomigrate"
	}

	config.MainMigrationFile = "migrator.go"

	pwd, err := os.Getwd()
	if err != nil {
		panic("Error getting pwd: " + err.Error())
	}
	config.ProjectPath = pwd

	baseDir := pwd
	fileStat, err := os.Stat(baseDir + "/src")
	if err == nil && fileStat.IsDir() {
		baseDir += "/src"
	}
	config.LocalMigratorPath = baseDir + "/migrator"
	config.LocalMigrationsPath = config.LocalMigratorPath + "/migrations"
	config.LocalMigrationsPackage = "migrations"
	config.LocalTemplatesPath = config.LocalMigratorPath + "/templates"

	return config
}

func LoadConfig() *Config {
	b, err := ReadFile(ConfigFileName)
	if err != nil {
		panic("Couldn't read " + ConfigFileName + ": " + err.Error())
	}
	config := &Config{}
	if err := json.Unmarshal(b, config); err != nil {
		panic("Couldn't parse " + ConfigFileName + " json: " + err.Error())
	}
	return config
}
