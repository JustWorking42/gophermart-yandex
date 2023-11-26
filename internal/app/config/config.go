package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAdress  string
	DatabaseURI   string
	AccrualAdress string
}

func ConfigService() (*Config, error) {

	serverConfig := Config{}
	os.Environ()
	flag.StringVar(&serverConfig.ServerAdress, "a", ":8081", "")
	flag.StringVar(&serverConfig.DatabaseURI, "d", "postgresql://postgres:2446Ba3dDc3AAB-bgd1Fe25f636A1e42@viaduct.proxy.rlwy.net:22501/railway", "")
	flag.StringVar(&serverConfig.AccrualAdress, "r", "http://127.0.0.1:8080", "")
	flag.Parse()

	if serverAdress := os.Getenv("RUN_ADDRESS"); serverAdress != "" {
		serverConfig.ServerAdress = serverAdress
	}

	if dbStoragePath, exist := os.LookupEnv("DATABASE_URI"); exist {
		serverConfig.DatabaseURI = dbStoragePath
	}

	if accrualAdress, exist := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exist {
		serverConfig.AccrualAdress = accrualAdress
	}

	return &serverConfig, nil

}
