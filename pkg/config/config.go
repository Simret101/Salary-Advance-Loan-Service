package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	JWTSecret     string 
	DBURL         string
	SMTPHost      string
	SMTPPort      int
	EmailFrom     string
	EmailPassword string
	AdminEmail    string
	AdminPassword string
	DBDsn         string
	PrivateKeyPEM string
	PublicKeyPEM  string
	Issuer        string
	AccessTTLMin  int
	RefreshTTLMin int
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	smtpPort, err := strconv.Atoi(getenv("SMTP_PORT", "587"))
	if err != nil {
		log.Fatal("Error converting SMTP_PORT to int:", err)
	}

	cfg := Config{
		AppPort:       getenv("APP_PORT", "8080"),
		JWTSecret:     getenv("JWT_SECRET", ""),
		DBURL:         getenv("DB_URL", ""),
		SMTPHost:      getenv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:      smtpPort,
		EmailFrom:     getenv("EMAIL_FROM", "semretb74@gmail.com"),
		EmailPassword: getenv("EMAIL_PASSWORD", ""),
		AdminEmail:    getenv("ADMIN_EMAIL", "semretb74@gmail.com"),
		AdminPassword: getenv("ADMIN_PASSWORD", "1234!!@#AR"),
		DBDsn:         getenv("DB_DSN", ""),
		PrivateKeyPEM: getenv("JWT_PRIVATE_KEY", ""),
		PublicKeyPEM:  getenv("JWT_PUBLIC_KEY", ""),
		Issuer:        getenv("JWT_ISSUER", "invite-rs256"),
		AccessTTLMin:  getenvInt("ACCESS_TTL_MIN", 15),
		RefreshTTLMin: getenvInt("REFRESH_TTL_MIN", 60*24*7),
	}

	log.Printf("issuer=%s access=%d refresh=%d", cfg.Issuer, cfg.AccessTTLMin, cfg.RefreshTTLMin)
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		_, _ = fmt.Sscanf(v, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return def
}
