package main

import (
	"dragon/api"
	"dragon/bot"
	"dragon/db"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		getenv("DB_USER", "root"),
		getenv("DB_PASSWORD", "12345"),
		getenv("DB_HOST", "127.0.0.1"),
		getenv("DB_PORT", "3306"),
		getenv("DB_NAME", "dragon"),
	)

	conn, err := db.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := db.EnsureSchema(conn); err != nil {
		log.Fatal("миграции: ", err)
	}
	if err := db.SyncCatalog(conn); err != nil {
		log.Println("seed:", err)
	}

	adminToken := getenv("ADMIN_TOKEN", "")
	if adminToken == "" {
		adminToken = getenv("TOKEN", "")
	}
	if adminToken == "" {
		adminToken = "dragon-admin-2026"
		log.Println("ВНИМАНИЕ: ADMIN_TOKEN не задан в .env — используется пароль по умолчанию:", adminToken)
	}

	siteURL := getenv("SITE_URL", "http://"+getenv("ADDR", "127.0.0.1:8080"))
	bot.StartReferralBot(conn, getenv("REF_BOT_TOKEN", ""), siteURL)
	adminBot := bot.StartAdminBot(conn, getenv("ADMIN_BOT_TOKEN", ""), getenv("ADMIN_BOT_SECRET", ""))

	addr := getenv("ADDR", "127.0.0.1:8080")

	log.Println("Слушаю", addr, "(за nginx)")
	log.Fatal(http.ListenAndServe(addr, api.Register(conn, adminToken, adminBot)))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
