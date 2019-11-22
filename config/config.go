package config

import (
	"os"
)

// Config struct with env file
type Config struct {
	GameList          string
	GameDisscounts    string
	GameSource        string
	GameIDPrefix      string
	BotKey            string
	BaseURL           string
	BaseURLSales      string
	BaseURLSalesParam string
}

// GetEnv .
var GetEnv Config

// New returns a new Config struct
func New() *Config {
	GetEnv = Config{
		GameList:          getEnvStr("GAME_LIST", ""),
		GameDisscounts:    getEnvStr("GAME_DISSCOUNTS", ""),
		GameSource:        getEnvStr("GAME_SOURCE", ""),
		GameIDPrefix:      getEnvStr("GAME_PREFIX", ""),
		BotKey:            getEnvStr("TELEGRAM_BOT_KEY", ""),
		BaseURL:           getEnvStr("BASE_URL_PS", ""),
		BaseURLSales:      getEnvStr("BASE_URL_SALES_PS", ""),
		BaseURLSalesParam: getEnvStr("BASE_URL_SALES_PARAM", ""),
	}

	return &GetEnv
}

// Simple helper function to read an environment or return a default value
func getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	// всегда возвращяет пустую строук нет необходимости использовать для этого переменную
	return defaultVal
}
