package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ssoroka/gomigrate/migrator"
)

var (
	installFlagSet       = flag.NewFlagSet("install", flag.PanicOnError)
	newMigrationFlagSet  = flag.NewFlagSet("new", flag.PanicOnError)
	upMigrationFlagSet   = flag.NewFlagSet("up", flag.PanicOnError)
	downMigrationFlagSet = flag.NewFlagSet("down", flag.PanicOnError)

	options = &migrator.Options{
		Install: migrator.InstallOptions{
			Help: installFlagSet.Bool("help", false, "Help"),
		},
		New: migrator.NewOptions{
			Name: newMigrationFlagSet.String("name", "", "A name for the migration. (Required)"),
			Help: newMigrationFlagSet.Bool("help", false, "Help"),
		},
		Build: migrator.BuildOptions{},
		Up: migrator.UpDownOptions{
			PreDeployOnly:  upMigrationFlagSet.Bool("pre", false, "Run Pre-deploy scripts only (default is all)"),
			PostDeployOnly: upMigrationFlagSet.Bool("post", false, "Run Post-deploy scripts only (default is all)"),
			Version:        upMigrationFlagSet.String("version", "", "Run up only on this version"),
			Force:          upMigrationFlagSet.Bool("force", false, "Force the migration to run, even if it has already run successfully"),
			Help:           upMigrationFlagSet.Bool("help", false, "Help"),
		},
		Down: migrator.UpDownOptions{
			PreDeployOnly:  downMigrationFlagSet.Bool("pre", false, "Run Pre-deploy scripts only (default is all)"),
			PostDeployOnly: downMigrationFlagSet.Bool("post", false, "Run Post-deploy scripts only (default is all)"),
			Version:        downMigrationFlagSet.String("version", "", "Run down only on this version"),
			Force:          downMigrationFlagSet.Bool("force", false, "Force the migration to run, even if it has not run, or already run down successfully"),
			Help:           downMigrationFlagSet.Bool("help", false, "Help"),
		},
	}

	help      = flag.Bool("help", false, "Get usage")
	usageText = `Usage of migrate:
	
		migrate install [-help]                Run once on a new project to install the default migration files
		migrate new [-help]                    Creates a new migration script in your project
		migrate up [-help]                     Runs all pending migrations
		migrate down [-help]                   Runs a migration down, used typically for a specific migration version
		migrate build                          Build the migrator binary used to run migrations in production without local dependencies on the go language
`
)

func main() {
	flag.Parse()

	if *help || len(os.Args) < 2 {
		fmt.Println(usageText)
		return
	}

	switch os.Args[1] {
	case "install":
		if err := installFlagSet.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		if *options.Install.Help {
			installFlagSet.Usage()
			os.Exit(2)
		}
		install(&options.Install)
	case "build":
		migrator.BuildMigrator()
	case "new":
		if err := newMigrationFlagSet.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		if *options.New.Help || *options.New.Name == "" {
			newMigrationFlagSet.Usage()
			os.Exit(2)
		}
		migrator.NewMigrationFile(&options.New)
	case "up":
		if err := upMigrationFlagSet.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		if *options.Up.Help {
			upMigrationFlagSet.Usage()
			os.Exit(2)
		}
		migrator.UpMigration(&options.Up)
	case "down":
		if err := downMigrationFlagSet.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
		if *options.Down.Help {
			downMigrationFlagSet.Usage()
			os.Exit(2)
		}
		migrator.DownMigration(&options.Down)
	default:
		fmt.Println(usageText)
		os.Exit(2)
	}

}
