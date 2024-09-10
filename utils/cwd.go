package utils

import "os"

func GetCWD() string {
	cwd, err := os.Getwd()
	if err == nil {
		return cwd
	}
	return "/"
}
