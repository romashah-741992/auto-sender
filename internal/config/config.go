package config

import "os"

type Config struct {
	Port        string
	DBDSN       string
	RedisAddr   string
	WebhookURL  string
	WebhookAuth string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DBDSN:       getEnv("DB_DSN", ""),           // e.g. appuser:apppass@tcp(db:3306)/auto_sender?parseTime=true&loc=UTC
		RedisAddr:   getEnv("REDIS_ADDR", ""),       // e.g. redis:6379
		WebhookURL:  getEnv("WEBHOOK_URL", ""),      // webhook.site URL
		WebhookAuth: getEnv("WEBHOOK_AUTH_KEY", ""), // x-ins-auth-key
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
