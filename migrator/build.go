package migrator

func BuildMigrator() {
	config = LoadConfig()

	buildMigrationBinary()
}
