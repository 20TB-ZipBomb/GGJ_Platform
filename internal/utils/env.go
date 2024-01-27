package utils

import (
	"flag"
)

// Evaluates if the server is currently running in production
func IsProductionEnv() bool {
	return SanitizeEnvFlag(flag.Lookup("env").Value.String()) == "prod"
}

// Ensures that the server runs in development if the environment command-line argument is malformed
func SanitizeEnvFlag(env string) string {
	if env != "dev" && env != "prod" {
		return "dev"
	}

	return env
}
