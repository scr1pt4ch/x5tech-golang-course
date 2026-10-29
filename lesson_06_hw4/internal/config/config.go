package config

import "os"

const DefaultHTTPAddr = ":8080"

type Config struct {
	HTTPAddr string
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")

	if "" == addr {
		addr = DefaultHTTPAddr
	}

	return Config{HTTPAddr: addr}
}
