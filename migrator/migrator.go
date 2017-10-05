package migrator

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Migrator struct {
	Migrations SortableMigrations
	DbDriver   DbDriver
}

type direction string
type scope string

const (
	scopePreMigration  scope = "pre"
	scopePostMigration       = "post"

	directionUp   direction = "up"
	directionDown           = "down"
)

var (
	options = &UpDownOptions{
		PreDeployOnly:  flag.Bool("pre", false, "Run Pre-deploy scripts only (default is all)"),
		PostDeployOnly: flag.Bool("post", false, "Run Post-deploy scripts only (default is all)"),
		Version:        flag.String("version", "", "Run up only on this version"),
		Force:          flag.Bool("force", false, "Force the migration to run, even if it has already run successfully"),
	}
	up   = flag.Bool("up", false, "Run up scripts")
	down = flag.Bool("down", false, "Run down scripts")
)

type DbDriver interface {
	GetAllRunVersions(scope string) ([]int64, error)
	InsertVersion(scope string, version int64) error
	RemoveVersion(scope string, version int64) error
}

func NewMigrator() *Migrator {

	return &Migrator{}
}

func (m *Migrator) PreMigration(f func())  {}
func (m *Migrator) PostMigration(f func()) {}
func (m *Migrator) PostFailure(f func())   {}

func (m *Migrator) Register(mig *Migration) {
	m.Migrations = append(m.Migrations, mig)
}

// Run is responsible for running all pending migrations
func (m *Migrator) Run() {
	flag.Parse()
	var err error
	sort.Sort(m.Migrations)

	runPre := true
	runPost := true
	if options.PreDeployOnly != nil && *options.PreDeployOnly {
		runPost = false
	}
	if options.PostDeployOnly != nil && *options.PostDeployOnly {
		runPre = false
	}
	if !runPre && !runPost {
		fmt.Println("Cannot use both -pre and -post exclusivity flags at the same time. If you want to run both (which is the default), don't supply -pre and -post arguments")
		os.Exit(2)
	}
	version := ""
	if options.Version != nil {
		version = strings.Replace(*options.Version, "-", "", -1)
		version = strings.Replace(version, `"`, "", -1)
	}
	force := options.Force != nil && *options.Force
	if force && version == "" {
		fmt.Println("Cannot use -force without -version")
		os.Exit(3)
	}
	if down != nil && *down && version == "" {
		fmt.Println("Cannot run down migrations without a version specified")
		os.Exit(6)
	}

	m.setRunStates()

	if (up != nil && *up) || down == nil || !*down {
		for _, mig := range m.Migrations {
			if version != "" && strconv.FormatInt(mig.OrderingNumber, 10) != version {
				continue
			}
			if runPre {
				if !mig.preHasRun || force {
					if err = m.runFunctionHook(mig, mig.upFunc, directionUp, scopePreMigration, mig.OrderingNumber); err != nil {
						m.runFunctionHook(mig, mig.downFunc, directionDown, scopePreMigration, mig.OrderingNumber)
						os.Exit(4)
					}
				}
			}
			if runPost {
				if !mig.postHasRun || force {
					if err = m.runFunctionHook(mig, mig.postUpFunc, directionUp, scopePostMigration, mig.OrderingNumber); err != nil {
						m.runFunctionHook(mig, mig.postDownFunc, directionDown, scopePostMigration, mig.OrderingNumber)
						if runPre {
							m.runFunctionHook(mig, mig.downFunc, directionDown, scopePreMigration, mig.OrderingNumber)
						}
						os.Exit(4)
					}
				}
			}
		}
	} else { // run down
		for i := len(m.Migrations) - 1; i >= 0; i-- {
			mig := m.Migrations[i]
			if version != "" && strconv.FormatInt(mig.OrderingNumber, 10) != version {
				continue
			}
			if runPost {
				if mig.postHasRun || force {
					if err = m.runFunctionHook(mig, mig.postDownFunc, directionDown, scopePostMigration, mig.OrderingNumber); err != nil {
						os.Exit(4)
					}
				}
			}
			if runPre {
				if mig.preHasRun || force {
					if err = m.runFunctionHook(mig, mig.downFunc, directionDown, scopePreMigration, mig.OrderingNumber); err != nil {
						os.Exit(4)
					}
				}
			}
		}
	}
}

func (m *Migrator) setRunStates() {
	preMigrationsRunInts, err := m.DbDriver.GetAllRunVersions(string(scopePreMigration))
	if err != nil {
		panic("Error getting run versions: " + err.Error())
	}
	postMigrationsRunInts, err := m.DbDriver.GetAllRunVersions(string(scopePostMigration))
	if err != nil {
		panic("Error getting run versions: " + err.Error())
	}
	preMigrationsRun := buildMapFromIntArray(preMigrationsRunInts)
	postMigrationsRun := buildMapFromIntArray(postMigrationsRunInts)

	for i := range m.Migrations {
		_, m.Migrations[i].preHasRun = preMigrationsRun[m.Migrations[i].OrderingNumber]
		_, m.Migrations[i].postHasRun = postMigrationsRun[m.Migrations[i].OrderingNumber]
	}
}

func (m *Migrator) runFunctionHook(mig *Migration, f migrationStepFunc, direction direction, scope scope, version int64) error {
	if f != nil {
		mig.Output("Running " + string(scope) + "-" + string(direction) + " migration")

		if err := f(m); err != nil {
			mig.Output(fmt.Sprintf("Failed to run %s-%s migration: %v", scope, direction, err))
			return err
		}
	}
	if direction == directionUp {
		if err := m.DbDriver.InsertVersion(string(scope), version); err != nil {
			mig.Output(fmt.Sprintf("Error Inserting version %d scope %s into db: %s", version, scope, err.Error()))
			return err
		}
	} else {
		if err := m.DbDriver.RemoveVersion(string(scope), version); err != nil {
			mig.Output(fmt.Sprintf("Error Removeing version %d scope %s from db: %s", version, scope, err.Error()))
			return err
		}
	}
	return nil
}

type SortableMigrations []*Migration

func (m SortableMigrations) Len() int {
	return len(m)
}

func (m SortableMigrations) Less(a, b int) bool {
	return m[a].OrderingNumber < m[b].OrderingNumber
}

func (m SortableMigrations) Swap(a, b int) {
	m[a], m[b] = m[b], m[a]
}

func buildMapFromIntArray(ints []int64) map[int64]struct{} {
	result := map[int64]struct{}{}
	for i := range ints {
		result[ints[i]] = struct{}{}
	}
	return result
}
