package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config type that stores 0tfile values.
type Config struct {
	// max upload size in MB
	MaxUploadSize int
	// max download count limit
	MaxDownloadCount int
	// path where files will be uploaded
	UploadPath string
	// limit of time for the file stored in days
	MaxUploadedDays int
	// base url used to return the download url and one time password
	BaseUrl string
}

// Returns a Config type with 0tfile values.
// The values are taken from the next sources in priority order.
// ENV
// .env file
// last resort defaults stated in the code
func Load() *Config {
	_ = godotenv.Load()

	// Right now this could fail if we set bad env variables

	maxUploadSize, _ := strconv.Atoi(getEnv("MAX_UPLOAD_SIZE", "100"))
	maxDownloadCount, _ := strconv.Atoi(getEnv("MAX_DOWNLOAD_COUNT", "3"))
	uploadPath := getEnv("UPLOAD_PATH", "/tmp/uploads")
	maxUploadedDays, _ := strconv.Atoi(getEnv("MAX_UPLOADED_DAYS", "14"))
	baseUrl := getEnv("BASE_URL", "localhost")

	return &Config{
		MaxUploadSize:    maxUploadSize,
		MaxDownloadCount: maxDownloadCount,
		UploadPath:       uploadPath,
		MaxUploadedDays:  maxUploadedDays,
		BaseUrl:          baseUrl,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return fallback
}
