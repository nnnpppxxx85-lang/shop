package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Settings struct {
	PaymentCard        string `json:"paymentCard"`
	PaymentPhone       string `json:"paymentPhone"`
	ConsultantTelegram string `json:"consultantTelegram"`
}

func (s *Server) settingsMap() map[string]string {
	out := map[string]string{}
	rows, err := s.DB.Query(`SELECT setting_key, value FROM settings`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			out[k] = v
		}
	}
	return out
}

// GET /api/settings — публичные настройки витрины: реквизиты для оплаты
// переводом и ссылка на консультанта в Telegram.
func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	m := s.settingsMap()
	writeJSON(w, Settings{
		PaymentCard:        m["payment_card"],
		PaymentPhone:       m["payment_phone"],
		ConsultantTelegram: m["consultant_telegram"],
	})
}

// PUT /api/admin/settings
func (s *Server) adminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var in Settings
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	pairs := map[string]string{
		"payment_card":        strings.TrimSpace(in.PaymentCard),
		"payment_phone":       strings.TrimSpace(in.PaymentPhone),
		"consultant_telegram": strings.TrimSpace(in.ConsultantTelegram),
	}
	for key, value := range pairs {
		if _, err := s.DB.Exec(
			`INSERT INTO settings (setting_key, value) VALUES (?,?)
			 ON DUPLICATE KEY UPDATE value = VALUES(value)`,
			key, value,
		); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	writeJSON(w, map[string]any{"ok": true})
}
