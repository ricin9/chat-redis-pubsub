package config

import "flag"

var (
	Port = flag.String("port", ":3000", "Port to listen on")
	Prod = flag.Bool("prod", false, "Production mode")
)
