package runner

import (
	"errors"
	"os"
)

func IsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}
