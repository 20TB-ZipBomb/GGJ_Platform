package utils

import(
    "os"
)

func IsProductionEnv() bool {
    return os.Getenv("SERVER_ENV") == "production"
}
