package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ssoroka/gomigrate/migrator"
)

var (
	defaultConfig *migrator.Config
	config        *migrator.Config
)

func init() {
	defaultConfig = migrator.DefaultConfig()
}

func install(options *migrator.InstallOptions) {
	fmt.Println("Checking for config ...")

	createDefaultConfigFile(options)

	config = migrator.LoadConfig()

	fmt.Println("Creating migration folder ...")
	createDefaultFiles(options)
	fmt.Println("\nYou're all set.\n\tCreate a migration with `migrate new`\n\tRun all migrations with `migrate`")
}

// createDefaultConfigFile creates a config file if one doesn't already exist
func createDefaultConfigFile(options *migrator.InstallOptions) {
	if migrator.FileExists(migrator.ConfigFileName) {
		return
	}
	f, err := os.Create(migrator.ConfigFileName)
	if err != nil {
		panic("Could not open file: " + err.Error())
	}

	serialized, err := json.Marshal(defaultConfig)
	if err != nil {
		panic("Could not serialize default config: " + err.Error())
	}

	_, err = f.Write(serialized)
	if err != nil {
		panic("Could not write to " + migrator.ConfigFileName + ": " + err.Error())
	}

	fmt.Println("Created default config file .migrate\nYou can optionally edit this file, then run `migrate install` again.")
	f.Close()

	os.Exit(1)
}

func createDefaultFiles(options *migrator.InstallOptions) {
	createDir(config.LocalMigratorPath)
	createDir(config.LocalMigrationsPath)
	createDir(config.LocalTemplatesPath)

	installFile("new_migration.tmpl", config.LocalTemplatesPath, "new_migration.tmpl")
	installFile("new_migrator_install.tmpl", config.LocalMigratorPath, config.MainMigrationFile)
}

func createDir(path string) {
	if err := os.Mkdir(path, 0755); err != nil {
		if os.IsExist(err) {
			return
		}
		panic("Couldn't create folder " + path + ": " + err.Error())
	}
}

func installFile(templateName, path, destName string) {
	if migrator.FileExists(filepath.Join(path, destName)) {
		return
	}

	content, err := migrator.ReadFile(filepath.Join(config.GoMigratePackagePath, "default_templates", templateName))
	if err != nil {
		panic("Could not read template " + filepath.Join(config.GoMigratePackagePath, "default_templates", templateName) + ": " + err.Error())
	}

	if err := migrator.WriteFile(filepath.Join(path, destName), content); err != nil {
		panic("could not write file " + filepath.Join(path, destName) + ": " + err.Error())
	}
}
