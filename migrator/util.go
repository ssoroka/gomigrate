package migrator

import (
	"os"
)

func FileExists(filename string) bool {
	f, err := os.Open(filename)
	if f != nil {
		f.Close()
	}
	return !os.IsNotExist(err)
}
