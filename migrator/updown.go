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
	if !migrationBinaryExists() {
		buildMigrationBinary()
	}
	runMigrationBinary(direction)

	fmt.Println("Done migrating")
}

func migrationBinaryExists() bool {
	return FileExists(migratorBinFile())
}

func buildMigrationBinary() {
	// go build -o binary source driver
	goArgs := []string{"build", "-o", migratorBinFile(), originMigratorGoFile()}

	driverFile := path.Join(config.LocalMigratorPath, "driver.go")
	if FileExists(driverFile) {
		goArgs = append(goArgs, driverFile)
	}

	fmt.Println("go", goArgs)
	buildCmd := exec.Command("go", goArgs...)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		panic("Couldn't build migration binary: " + err.Error())
	}
}

func runMigrationBinary(direction string) {
	// migratorBinary -config etc
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

	bin := migratorBinFile()
	fmt.Println(bin, migratorArgs)

	cmd := exec.Command(bin, migratorArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic("Couldn't run migrations: " + err.Error())
	}

}

func originMigratorGoFile() string {
	return path.Join(config.LocalMigratorPath, config.MainMigrationFile)
}

func migratorBinFile() string {
	s := originMigratorGoFile()
	return s[0 : len(s)-3]
}
