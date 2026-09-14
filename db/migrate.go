package db

import (
	"database/sql"
	"fmt"
	"os"
)

// EnsureSchema применяет базовые SQL-файлы (idempotent благодаря IF NOT EXISTS)
// и докатывает точечные изменения схемы, появившиеся позже, не трогая
// уже существующие таблицы и данные.
func EnsureSchema(conn *sql.DB) error {
	for _, file := range []string{"001_init.sql", "002_shop.sql"} {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("чтение %s: %w", file, err)
		}
		if _, err := conn.Exec(string(content)); err != nil {
			return fmt.Errorf("применение %s: %w", file, err)
		}
	}

	columns := []struct{ table, column, ddl string }{
		{"products", "discount_type", "ENUM('none','percent','bundle') NOT NULL DEFAULT 'none'"},
		{"products", "bundle_buy_qty", "TINYINT UNSIGNED DEFAULT NULL"},
		{"products", "bundle_total_qty", "TINYINT UNSIGNED DEFAULT NULL"},
		{"order_items", "line_total", "INT UNSIGNED NOT NULL DEFAULT 0"},
		{"orders", "referral_partner_id", "INT UNSIGNED DEFAULT NULL"},
		{"orders", "referral_username", "VARCHAR(191) DEFAULT NULL"},
		{"orders", "receipt_path", "VARCHAR(255) DEFAULT NULL"},
		{"products", "storage_options", "TEXT DEFAULT NULL"},
		{"orders", "access_token", "VARCHAR(64) DEFAULT NULL UNIQUE"},
	}
	for _, c := range columns {
		if err := addColumnIfMissing(conn, c.table, c.column, c.ddl); err != nil {
			return fmt.Errorf("ALTER %s.%s: %w", c.table, c.column, err)
		}
	}

	// разовый бэкфилл для записей, созданных до появления новых колонок
	if _, err := conn.Exec(`UPDATE products SET discount_type = 'percent' WHERE discount_percent > 0 AND discount_type = 'none'`); err != nil {
		return err
	}
	if _, err := conn.Exec(`UPDATE order_items SET line_total = price * qty WHERE line_total = 0`); err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			setting_key VARCHAR(64) PRIMARY KEY,
			value VARCHAR(255) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS partners (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			telegram_id BIGINT NOT NULL UNIQUE,
			username VARCHAR(191) DEFAULT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS admin_recipients (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			chat_id BIGINT NOT NULL UNIQUE,
			username VARCHAR(191) DEFAULT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS contact_messages (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(191) DEFAULT NULL,
			phone VARCHAR(64) DEFAULT NULL,
			message TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, ddl := range tables {
		if _, err := conn.Exec(ddl); err != nil {
			return fmt.Errorf("создание таблицы: %w", err)
		}
	}

	// payment_card/payment_phone нарочно пустые: показывать покупателю
	// придуманный номер карты до того, как админ впишет реальные
	// реквизиты в настройках, опаснее, чем пустое поле.
	defaults := map[string]string{
		"payment_card":        "",
		"payment_phone":       "",
		"payment_method":      "card",
		"consultant_telegram": "https://t.me/dragonmobile_support",
	}
	for key, value := range defaults {
		if _, err := conn.Exec(`INSERT IGNORE INTO settings (setting_key, value) VALUES (?, ?)`, key, value); err != nil {
			return fmt.Errorf("настройка %s: %w", key, err)
		}
	}

	return nil
}

func addColumnIfMissing(conn *sql.DB, table, column, ddl string) error {
	var count int
	err := conn.QueryRow(
		`SELECT COUNT(*) FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table, column,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err = conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, ddl))
	return err
}
