package config

// tag::import[]
import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

/**
 * ReadConfig reads the application settings from enviroment varibales - its possible to use .env file
 */

func LoadConfiguraion() (*Config, error) {
	config := Config{}

	config.JwtSecret = os.Getenv("APP_JWT_SECRET")
	// config.Port = os.Getenv("APP_PORT")
	// config.SaltRounds = parseIntWithDefaultValue(os.Getenv("APP_BCRYPT_SALT_ROUNDS"), 12)

	return &config, nil
}

func parseIntWithDefaultValue(inputString string, defaultValue int32) int {
	result, err := strconv.ParseInt(inputString, 10, 32)

	if err != nil {
		result = int64(defaultValue)
	}

	return int(result)
}

type Config struct {
	Port       string
	JwtSecret  string
	SaltRounds int
}
