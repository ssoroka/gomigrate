package migrator

import "fmt"

type Migration struct {
	OrderingNumber  int64
	FormattedNumber string
	preHasRun       bool
	postHasRun      bool
	upFunc          migrationStepFunc
	downFunc        migrationStepFunc
	postUpFunc      migrationStepFunc
	postDownFunc    migrationStepFunc
	verifyFunc      migrationStepFunc
}

type migrationStepFunc func(migrator *Migrator) error

func NewMigration(number int64) *Migration {
	m := &Migration{
		OrderingNumber: number,
	}
	numberStr := fmt.Sprintf("%d", number)
	yearStr := numberStr[0:4]
	monthStr := numberStr[4:6]
	dayStr := numberStr[6:8]
	secondsStr := numberStr[8:13]
	m.FormattedNumber = yearStr + "_" + monthStr + "_" + dayStr + "_" + secondsStr
	// year, err := strconv.ParseInt(yearStr, 10, 32)
	// if err != nil {
	// 	panic(err)
	// }
	// month, err := strconv.ParseInt(monthStr, 10, 32)
	// if err != nil {
	// 	panic(err)
	// }
	// day, err := strconv.ParseInt(dayStr, 10, 32)
	// if err != nil {
	// 	panic(err)
	// }
	// seconds, err := strconv.ParseInt(secondsStr, 10, 32)
	// if err != nil {
	// 	panic(err)
	// }
	// hours := seconds / time.Hour
	// seconds = seconds % time.Hour
	// minutes := seconds / time.Minute
	// seconds = seconds % time.Minute

	// m.Timestamp = time.Date(year, month, day, hours, minutes, seconds, 0, time.Local())
	return m
}

func (m *Migration) Up(f migrationStepFunc) *Migration {
	m.upFunc = f
	return m
}

func (m *Migration) Down(f migrationStepFunc) *Migration {
	m.downFunc = f
	return m
}

func (m *Migration) PostUp(f migrationStepFunc) *Migration {
	m.postUpFunc = f
	return m
}

func (m *Migration) PostDown(f migrationStepFunc) *Migration {
	m.postDownFunc = f
	return m
}

func (m *Migration) Verify(f migrationStepFunc) *Migration {
	m.verifyFunc = f
	return m
}

func (m *Migration) Output(s string) {
	fmt.Println("[" + m.FormattedNumber + "] " + s)
}
