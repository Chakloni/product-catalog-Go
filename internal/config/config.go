package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	MongoDB  string
	Port     string
}

func LoadConfig() *Config {
	// Solo cargar .env en desarrollo local
	// En producción (Render) esto se ignora automáticamente
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ Error loading .env file:", err)
		} else {
			log.Println("✅ .env file loaded successfully")
		}
	} else {
		log.Println("🌐 Using system environment variables")
	}

	return &Config{
		MongoURI: getEnv("MONGO_URI", ""),
		MongoDB:  getEnv("MONGO_DB", "productCatalog"),
		Port:     getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}