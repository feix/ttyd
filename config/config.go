package config

import (
	"os"
)

func GetEnv(key string, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return defaultVal
}
