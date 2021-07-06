package testutils

import "github.com/rs/zerolog"

func SetLogLevel() {
	// Log level set to info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
