package utils

import "os"

// Evaluates if the server is currently running in production
func IsProductionEnv() bool {
	return os.Getenv("SERVER_ENV") == "production"
}
