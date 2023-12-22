package config

type config struct {
	DB    DBConfig
	Redis RedisConfig
}

type DBConfig struct {
	DNS string
}

type RedisConfig struct {
	Addr string
}
