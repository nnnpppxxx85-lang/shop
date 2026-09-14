package api

import (
	"crypto/rand"
	"database/sql"
	"dragon/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// newOrderToken генерирует непредсказуемую ссылку на заказ для покупателя:
// заказ создаётся анонимно (без логина), поэтому публичные ручки ищут его
// по этому токену, а не по номеру id — иначе подбором id можно было бы
// прочитать имя, телефон и состав чужого заказа.
func newOrderToken() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type orderItemInput struct {
	ProductID    int    `json:"productId"`
	Qty          int    `json:"qty"`
	StorageLabel string `json:"storageLabel"`
}

type orderInput struct {
	Name         string           `json:"name"`
	Phone        string           `json:"phone"`
	Email        string           `json:"email"`
	Address      string           `json:"address"`
	Comment      string           `json:"comment"`
	ReferralCode string           `json:"referralCode"`
	Items        []orderItemInput `json:"items"`
}

// resolveReferral проверяет код из ссылки (telegram_id партнёра) и
// возвращает его id в таблице partners и юзернейм (снимок на момент
// заказа — на случай, если партнёр потом сменит username в Telegram).
func (s *Server) resolveReferral(code string) (partnerID any, username any) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}
	tgID, err := strconv.ParseInt(code, 10, 64)
	if err != nil {
		return nil, nil
	}
	var id int
	var uname sql.NullString
	if err := s.DB.QueryRow(`SELECT id, username FROM partners WHERE telegram_id = ?`, tgID).Scan(&id, &uname); err != nil {
		return nil, nil
	}
	if uname.Valid && uname.String != "" {
		return id, uname.String
	}
	return id, nil
}

// POST /api/orders — создаёт заказ, цены берём из БД
func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var in orderInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if in.Name == "" || in.Phone == "" || len(in.Items) == 0 {
		http.Error(w, "заполните имя, телефон и добавьте товары", 400)
		return
	}

	referralPartnerID, referralUsername := s.resolveReferral(in.ReferralCode)

	token, err := newOrderToken()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	res, err := s.DB.Exec(`INSERT INTO orders
		(customer_name, phone, email, address, comment, total, payment_method, status, referral_partner_id, referral_username, access_token)
		VALUES (?,?,?,?,?,0,'transfer','new',?,?,?)`,
		in.Name, in.Phone, nullStr(in.Email), nullStr(in.Address), nullStr(in.Comment), referralPartnerID, referralUsername, token)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	orderID, _ := res.LastInsertId()

	total := 0
	for _, it := range in.Items {
		if it.Qty < 1 {
			it.Qty = 1
		}
		var name, discountType string
		var price, discountPercent int
		var bundleBuy, bundleTotal *int
		var rawStorage sql.NullString
		err := s.DB.QueryRow(`SELECT name, price, discount_type, discount_percent, bundle_buy_qty, bundle_total_qty, storage_options
			FROM products WHERE id = ?`, it.ProductID).
			Scan(&name, &price, &discountType, &discountPercent, &bundleBuy, &bundleTotal, &rawStorage)
		if err != nil {
			continue
		}
		fp := finalPrice(price, discountType, discountPercent)

		// доплату за объём памяти берём из сохранённых в товаре вариантов,
		// а не из запроса — иначе покупатель мог бы прислать любую скидку
		if label := strings.TrimSpace(it.StorageLabel); label != "" && rawStorage.Valid {
			for _, o := range db.DecodeStorageOptions(&rawStorage.String) {
				if o.Label == label {
					fp += o.PriceDelta
					name = name + " · " + o.Label
					break
				}
			}
		}

		payableQty := it.Qty
		if discountType == "bundle" {
			payableQty = bundlePayableQty(it.Qty, bundleBuy, bundleTotal)
		}
		lineTotal := fp * payableQty
		total += lineTotal
		_, _ = s.DB.Exec(`INSERT INTO order_items (order_id, product_id, name, price, qty, line_total) VALUES (?,?,?,?,?,?)`,
			orderID, it.ProductID, name, fp, it.Qty, lineTotal)
	}

	_, _ = s.DB.Exec(`UPDATE orders SET total = ? WHERE id = ?`, total, orderID)
	writeJSON(w, map[string]any{"ok": true, "token": token, "total": total})
}

// GET /api/orders/{ref} — публичная ручка для страницы оплаты. Ищем заказ
// по непредсказуемому access_token, а не по числовому id, чтобы чужой
// заказ нельзя было прочитать перебором номеров.
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")
	if ref == "" {
		http.NotFound(w, r)
		return
	}
	var id int
	var name, phone, status, method string
	var total int
	err := s.DB.QueryRow(`SELECT id, customer_name, phone, total, payment_method, status FROM orders WHERE access_token = ?`, ref).
		Scan(&id, &name, &phone, &total, &method, &status)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items := []map[string]any{}
	rows, err := s.DB.Query(`SELECT name, price, qty, line_total FROM order_items WHERE order_id = ?`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n string
			var price, qty, lineTotal int
			_ = rows.Scan(&n, &price, &qty, &lineTotal)
			items = append(items, map[string]any{"name": n, "price": price, "qty": qty, "lineTotal": lineTotal})
		}
	}
	writeJSON(w, map[string]any{"id": id, "name": name, "phone": phone, "total": total,
		"paymentMethod": method, "status": status, "items": items})
}

// POST /api/orders/{ref}/confirm-payment — покупатель прикладывает PDF
// квитанцию о переводе; заказ переходит в статус "ждёт проверки
// менеджером", а квитанция и данные заказа уходят в админ-бот. Заказ, как
// и в getOrder, ищем по access_token, а не по числовому id.
func (s *Server) confirmPayment(w http.ResponseWriter, r *http.Request) {
	ref := r.PathValue("ref")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "файл слишком большой (максимум 10 МБ)", 400)
		return
	}
	file, header, err := r.FormFile("receipt")
	if err != nil {
		http.Error(w, "прикрепите квитанцию в формате PDF", 400)
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		http.Error(w, "квитанция должна быть в формате PDF", 400)
		return
	}

	var id int
	if err := s.DB.QueryRow(`SELECT id FROM orders WHERE access_token = ?`, ref).Scan(&id); err != nil {
		http.NotFound(w, r)
		return
	}

	if err := os.MkdirAll("uploads/receipts", 0o755); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	receiptPath := fmt.Sprintf("uploads/receipts/order-%d.pdf", id)
	out, err := os.Create(receiptPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		http.Error(w, err.Error(), 500)
		return
	}
	out.Close()

	if _, err := s.DB.Exec(`UPDATE orders SET status='awaiting_confirmation', receipt_path=? WHERE id=?`, receiptPath, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.notifyAdminsOfOrder(id, receiptPath)
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) notifyAdminsOfOrder(orderID int, receiptPath string) {
	var name, phone, address, referralUsername string
	var total int
	err := s.DB.QueryRow(
		`SELECT customer_name, phone, COALESCE(address,''), total, COALESCE(referral_username,'') FROM orders WHERE id=?`,
		orderID,
	).Scan(&name, &phone, &address, &total, &referralUsername)
	if err != nil {
		return
	}

	var itemsText strings.Builder
	rows, err := s.DB.Query(`SELECT name, qty, line_total FROM order_items WHERE order_id=?`, orderID)
	if err == nil {
		for rows.Next() {
			var n string
			var qty, lineTotal int
			if rows.Scan(&n, &qty, &lineTotal) == nil {
				fmt.Fprintf(&itemsText, "• %s × %d\n", n, qty)
			}
		}
		rows.Close()
	}

	s.AdminBot.NotifyOrder(orderID, name, phone, address, total, itemsText.String(), referralUsername, receiptPath)
}

// PUT /api/admin/orders/{id}/status
func (s *Server) adminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var in struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	valid := map[string]bool{"new": true, "awaiting_confirmation": true, "paid": true, "cancelled": true}
	if !valid[in.Status] {
		http.Error(w, "неизвестный статус", 400)
		return
	}
	if _, err := s.DB.Exec(`UPDATE orders SET status=? WHERE id=?`, in.Status, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// GET /api/admin/orders/{id}/receipt — отдаёт PDF-квитанцию для просмотра в админке
func (s *Server) adminOrderReceipt(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var path sql.NullString
	if err := s.DB.QueryRow(`SELECT receipt_path FROM orders WHERE id=?`, id).Scan(&path); err != nil || !path.Valid || path.String == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path.String)
}
