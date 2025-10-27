package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
		return
	}
	log.Println("Файл .env успешно загружен")
}

func getString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func GetPort() string {
	return getString("PORT", ":1111")
}

func GetPath(key string) string {
	return getString(key, "tmp")
}

func GetApiVersion() string {
	return getString("FERZL_API_VER", "24.1.2")
}

func GetToken() string {
	return getString("FERZL_TOKEN", "TEST")
}

func GetTestEnp() string {
	return getString("TEST_ENP", "111111111111111")
}

/*
	func getBool(key string, defaultValue bool) bool {
		value := os.Getenv(key)
		i, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return i
	}
*/
type DatabaseConfig struct {
	Url string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Url: getString("DSN", "sqlserver://user:pswd@server-2:1433?database=PlanCon&connection+timeout=300"),
	}
}

type LogConfig struct {
	Level  int
	Format string
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  getInt("LOG_LEVEL", 0),
		Format: getString("LOG_FORMAT", "json"),
	}
}
