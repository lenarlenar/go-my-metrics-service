package flags

import "time"

const (
	DefaultServerAddress    = "localhost:8080"
	DefaultStoreIntervalSec = 300
	DefaultFileStoragePath  = ""  //"metrics.json"
	DefaultRestore          = true
	DefaultDatabaseDSN      = "" //"host=localhost port=5432 user=postgres password=admin dbname=postgres sslmode=disable"
	DefaultKey              = "key"
)

type Config struct {
	ServerAddress   string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	Key             string
}

type EnvConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}
