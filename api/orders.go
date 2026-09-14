package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type orderItemInput struct {
	ProductID int `json:"productId"`
	Qty       int `json:"qty"`
}

type orderInput struct {
	Name    string           `json:"name"`
	Phone   string           `json:"phone"`
	Email   string           `json:"email"`
	Address string           `json:"address"`
	Comment string           `json:"comment"`
	Items   []orderItemInput `json:"items"`
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

	res, err := s.DB.Exec(`INSERT INTO orders (customer_name, phone, email, address, comment, total, payment_method, status)
		VALUES (?,?,?,?,?,0,'sbp','new')`, in.Name, in.Phone, nullStr(in.Email), nullStr(in.Address), nullStr(in.Comment))
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
		err := s.DB.QueryRow(`SELECT name, price, discount_type, discount_percent, bundle_buy_qty, bundle_total_qty
			FROM products WHERE id = ?`, it.ProductID).
			Scan(&name, &price, &discountType, &discountPercent, &bundleBuy, &bundleTotal)
		if err != nil {
			continue
		}
		fp := finalPrice(price, discountType, discountPercent)
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
	writeJSON(w, map[string]any{"ok": true, "orderId": orderID, "total": total})
}

// GET /api/orders/{id}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var name, phone, status, method string
	var total int
	err := s.DB.QueryRow(`SELECT customer_name, phone, total, payment_method, status FROM orders WHERE id = ?`, id).
		Scan(&name, &phone, &total, &method, &status)
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

// POST /api/orders/{id}/pay — доступен только СБП
func (s *Server) payOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var body struct {
		Method string `json:"method"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.Method != "sbp" {
		w.WriteHeader(400)
		writeJSON(w, map[string]any{"ok": false, "error": "Оплата картой временно недоступна. Выберите СБП."})
		return
	}

	if _, err := s.DB.Exec(`UPDATE orders SET payment_method='sbp', status='awaiting_payment' WHERE id = ?`, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var total int
	_ = s.DB.QueryRow(`SELECT total FROM orders WHERE id = ?`, id).Scan(&total)

	// Демо-ссылка/QR СБП (в реальном проекте — ответ банка)
	payload := "https://qr.nspk.ru/demo-order-" + strconv.Itoa(id)
	writeJSON(w, map[string]any{
		"ok":      true,
		"method":  "sbp",
		"total":   total,
		"payload": payload,
		"qr":      "https://api.qrserver.com/v1/create-qr-code/?size=320x320&data=" + payload,
	})
}

// POST /api/orders/{id}/confirm — пометить как оплаченный (демо)
func (s *Server) confirmOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if _, err := s.DB.Exec(`UPDATE orders SET status='paid' WHERE id = ?`, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}
