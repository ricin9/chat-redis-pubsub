package config

import (
	"flag"
	"os"
)

var (
	Port string
	Prod = flag.Bool("prod", false, "Production mode")
)

func init() {
	port := os.Getenv("PORT")
	if port == "" {
		Port = ":3000"
	} else {
		Port = ":" + port
	}
}
