package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// requireAdmin защищает мутирующие ручки и данные с персональной
// информацией покупателей (заказы) простым токеном из .env — без него
// раньше /api/admin/* были полностью открыты кому угодно.
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.checkAdminToken(r) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func (s *Server) checkAdminToken(r *http.Request) bool {
	token := r.Header.Get("X-Admin-Token")
	if token == "" {
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if token == "" || s.AdminToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.AdminToken)) == 1
}

// POST /api/admin/login — проверяет пароль и возвращает токен для
// последующих запросов к /api/admin/*.
func (s *Server) adminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	if body.Password == "" || s.AdminToken == "" ||
		subtle.ConstantTimeCompare([]byte(body.Password), []byte(s.AdminToken)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, map[string]any{"ok": false, "error": "Неверный пароль"})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "token": s.AdminToken})
}
