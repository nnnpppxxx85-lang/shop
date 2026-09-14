package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type contactInput struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

// POST /api/contact — форма обратной связи со страницы «Контакты».
// Сообщение сохраняется и сразу уходит администраторам в Telegram.
func (s *Server) submitContact(w http.ResponseWriter, r *http.Request) {
	var in contactInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Message = strings.TrimSpace(in.Message)
	if in.Name == "" || in.Phone == "" {
		http.Error(w, "укажите имя и телефон", 400)
		return
	}

	if _, err := s.DB.Exec(
		`INSERT INTO contact_messages (name, phone, message) VALUES (?,?,?)`,
		in.Name, in.Phone, in.Message,
	); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.AdminBot.NotifyContact(in.Name, in.Phone, in.Message)
	writeJSON(w, map[string]any{"ok": true})
}
