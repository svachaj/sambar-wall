package config

// tag::import[]
import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

// Loads the application configuration from the environment variables.
func LoadConfiguraion() (*Config, error) {
	config := Config{}

	config.AppSecret = os.Getenv("APP_SECRET")
	config.AppPort = parseIntWithDefaultValue(os.Getenv("APP_PORT"), 5500)
	config.AppEnv = os.Getenv("APP_ENV")
	config.AppApplicationFormEmailCopy = os.Getenv("APP_APPLICATION_EMAIL_COPY")
	config.AppAccountIBAN = os.Getenv("APP_ACCOUNT_IBAN")
	config.AppAccountNumber = os.Getenv("APP_ACCOUNT_NUMBER")
	config.AppGeneratePaymentInfo = os.Getenv("APP_GENERATE_PAYMENT_INFO")
	config.AppCryptoKey = os.Getenv("APP_CRYPTO_KEY")

	// we assume that the database is MS SQL Server to backward compatibility
	// if we want to support other databases, we can simply change the database driver and connection string
	config.DbHost = os.Getenv("DB_HOST")
	config.DbPort = parseIntWithDefaultValue(os.Getenv("DB_PORT"), 1433)
	config.DbUser = os.Getenv("DB_USER")
	config.DbPassword = os.Getenv("DB_PASSWORD")
	config.DbName = os.Getenv("DB_NAME")

	config.SmtpHost = os.Getenv("SMTP_HOST")
	config.SmtpPort = parseIntWithDefaultValue(os.Getenv("SMTP_PORT"), 587)
	config.SmtpUser = os.Getenv("SMTP_USER")
	config.SmtpPassword = os.Getenv("SMTP_PASSWORD")

	config.StartPaymentsCheckingService = parseBoolWithDefault(os.Getenv("PAYMENTS_CRON"), false)

	config.ImapAddress = os.Getenv("IMAP_ADDRESS")
	config.ImapUsername = os.Getenv("IMAP_USERNAME")
	config.ImapPassword = os.Getenv("IMAP_PASSWORD")

	return &config, nil
}

func parseIntWithDefaultValue(inputString string, defaultValue int) (result int) {
	pint, err := strconv.ParseInt(inputString, 10, 32)

	if err != nil {
		result = defaultValue
	} else {
		result = int(pint)
	}

	return result
}

func parseBoolWithDefault(inputString string, defaultValue bool) (result bool) {

	if inputString == "1" {
		result = true
	} else {
		result = defaultValue
	}

	return result
}

// Holds the application settings.
type Config struct {
	AppPort                     int
	AppSecret                   string
	AppEnv                      string
	AppApplicationFormEmailCopy string
	AppAccountIBAN              string
	AppAccountNumber            string
	AppGeneratePaymentInfo      string
	AppCryptoKey                string

	// Database
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string

	// SMTP
	SmtpHost     string
	SmtpPort     int
	SmtpUser     string
	SmtpPassword string

	StartPaymentsCheckingService bool

	ImapAddress  string
	ImapUsername string
	ImapPassword string
}

const APP_ENV_PRODUCTION = "" // empty string or other string then localhost or test means production
const APP_ENV_LOCALHOST = "localhost"
const APP_ENV_TEST = "test"
