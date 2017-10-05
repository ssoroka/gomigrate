package migrator

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func UpMigration(options *UpDownOptions) {
	runMigration("up", options)
}

func DownMigration(options *UpDownOptions) {
	runMigration("down", options)
}

func runMigration(direction string, options *UpDownOptions) {
	config = LoadConfig()

	originMigratorGo := path.Join(config.LocalMigratorPath, config.MainMigrationFile)

	migratorArgs := []string{"-" + direction}
	if options.PostDeployOnly != nil && *options.PostDeployOnly {
		migratorArgs = append(migratorArgs, "-post")
	} else if options.PreDeployOnly != nil && *options.PreDeployOnly {
		migratorArgs = append(migratorArgs, "-pre")
	}
	if options.Force != nil && *options.Force {
		migratorArgs = append(migratorArgs, "-force")
	}
	if options.Version != nil && *options.Version != "" {
		migratorArgs = append(migratorArgs, "-version=\""+*options.Version+"\"")
	}

	for i := range os.Args {
		if os.Args[i] == "--" {
			migratorArgs = append(migratorArgs, os.Args[i+1:len(os.Args)]...)
		}
	}

	goArgs := []string{"run", originMigratorGo}

	driverFile := path.Join(config.LocalMigratorPath, "driver.go")
	if FileExists(driverFile) {
		goArgs = append(goArgs, driverFile)
	}

	goArgs = append(goArgs, migratorArgs...)
	fmt.Println("go", goArgs)
	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic("Couldn't run migrations: " + err.Error())
	}
	fmt.Println("Done migrating")
}
