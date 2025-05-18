package util

import (
	"os"
)

func GetEnv(name, fallback string) string {
	value, exists := os.LookupEnv(name)
	if !exists || len(value) == 0 {
		return fallback
	}
	return value
}
