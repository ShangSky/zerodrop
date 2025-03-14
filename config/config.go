package config

type Config struct {
	Addr string
}

func New(addr string) *Config {
	return &Config{Addr: addr}
}
