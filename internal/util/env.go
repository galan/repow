package util

import (
	"os"
)

func GetEnv(name, fallback string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}
	return value
}
