package config

import (
	"flag"

	"github.com/go-playground/validator/v10"
)

var (
	Validate *validator.Validate
)

func Setup() {
	flag.Parse()

	SetupDatabasesConfig()

	Validate = validator.New(validator.WithRequiredStructEnabled())
}
