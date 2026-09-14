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
