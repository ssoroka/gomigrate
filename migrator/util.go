package migrator

import (
	"os"
)

func FileExists(filename string) bool {
	f, err := os.Open(filename)
	if err == nil {
		f.Close()
		return true
	}
	return false
}
