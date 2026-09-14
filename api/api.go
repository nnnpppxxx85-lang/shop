package api

import (
	"database/sql"
	"dragon/bot"
	"encoding/json"
	"net/http"
)

type Server struct {
	DB         *sql.DB
	AdminToken string
	AdminBot   *bot.AdminBot
}

func Register(db *sql.DB, adminToken string, adminBot *bot.AdminBot) *http.ServeMux {
	s := &Server{DB: db, AdminToken: adminToken, AdminBot: adminBot}
	mux := http.NewServeMux()

	// статика
	mux.Handle("GET /assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir("templates/assets")),
	))

	// страницы
	mux.HandleFunc("GET /{$}", s.page("index.html"))
	mux.HandleFunc("GET /catalog", s.page("catalog.html"))
	mux.HandleFunc("GET /item", s.page("item.html"))
	mux.HandleFunc("GET /contacts", s.page("contacts.html"))
	mux.HandleFunc("GET /delivery", s.page("delivery.html"))
	mux.HandleFunc("GET /admin", s.page("admin.html"))
	mux.HandleFunc("GET /cart", s.page("cart.html"))
	mux.HandleFunc("GET /checkout", s.page("checkout.html"))
	mux.HandleFunc("GET /payment", s.page("payment.html"))

	// API
	mux.HandleFunc("GET /api/categories", s.categories)
	mux.HandleFunc("GET /api/catalog", s.catalog)
	mux.HandleFunc("GET /api/products/{slug}", s.product)
	mux.HandleFunc("GET /api/settings", s.settings)
	mux.HandleFunc("POST /api/contact", s.submitContact)

	// Заказы и оплата
	mux.HandleFunc("POST /api/orders", s.createOrder)
	mux.HandleFunc("GET /api/orders/{id}", s.getOrder)
	mux.HandleFunc("POST /api/orders/{id}/confirm-payment", s.confirmPayment)

	// Вход в админку
	mux.HandleFunc("POST /api/admin/login", s.adminLogin)

	// Админка — товары
	mux.HandleFunc("GET /api/admin/products", s.requireAdmin(s.adminProducts))
	mux.HandleFunc("POST /api/admin/products", s.requireAdmin(s.adminCreateProduct))
	mux.HandleFunc("PUT /api/admin/products/{id}", s.requireAdmin(s.adminUpdateProduct))
	mux.HandleFunc("DELETE /api/admin/products/{id}", s.requireAdmin(s.adminDeleteProduct))

	// Админка — категории
	mux.HandleFunc("POST /api/admin/categories", s.requireAdmin(s.adminCreateCategory))
	mux.HandleFunc("PUT /api/admin/categories/{slug}", s.requireAdmin(s.adminUpdateCategory))
	mux.HandleFunc("DELETE /api/admin/categories/{slug}", s.requireAdmin(s.adminDeleteCategory))

	// Админка — фото из templates/assets
	mux.HandleFunc("GET /api/admin/assets", s.requireAdmin(s.adminListAssetFolders))
	mux.HandleFunc("GET /api/admin/assets/{folder}", s.requireAdmin(s.adminFolderImages))

	// Админка — заказы
	mux.HandleFunc("GET /api/admin/orders", s.requireAdmin(s.adminOrders))
	mux.HandleFunc("PUT /api/admin/orders/{id}/status", s.requireAdmin(s.adminUpdateOrderStatus))
	mux.HandleFunc("GET /api/admin/orders/{id}/receipt", s.requireAdmin(s.adminOrderReceipt))

	// Админка — настройки (реквизиты оплаты, ссылка на консультанта)
	mux.HandleFunc("PUT /api/admin/settings", s.requireAdmin(s.adminUpdateSettings))

	return mux
}

func (s *Server) page(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/"+name)
	}
}

// ---------- модели ----------

type Category struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Code string `json:"code"`
	Note string `json:"note"`
}

type Product struct {
	ID              int      `json:"id"`
	Slug            string   `json:"slug"`
	Folder          string   `json:"folder"`
	Name            string   `json:"name"`
	Variant         *string  `json:"variant"`
	Price           int      `json:"price"`
	MarketPrice     *int     `json:"marketPrice"`
	DiscountType    string   `json:"discountType"`
	DiscountPercent int      `json:"discountPercent"`
	BundleBuyQty    *int     `json:"bundleBuyQty"`
	BundleTotalQty  *int     `json:"bundleTotalQty"`
	FinalPrice      int      `json:"finalPrice"`
	Source          *string  `json:"source"`
	CategorySlug    string   `json:"categorySlug"`
	CategoryName    string   `json:"categoryName"`
	Images          []string `json:"images"`
}

// finalPrice — цена за единицу товара. Для "комплектной" скидки (напр. 2+1=3)
// цена одной штуки не меняется — выгода считается при наборе нужного
// количества, см. bundlePayableQty.
func finalPrice(price int, discountType string, discountPercent int) int {
	if discountType == "percent" && discountPercent > 0 {
		return price - price*discountPercent/100
	}
	return price
}

// bundlePayableQty считает, за сколько единиц товара нужно заплатить при
// покупке qty штук с условием "купи buy — получи total". Например, для
// 2+1=3 (buy=2, total=3) при покупке 4 штук платим за 3 (одна группа из
// трёх даёт скидку, плюс один товар без скидки).
func bundlePayableQty(qty int, buy, total *int) int {
	if buy == nil || total == nil || *buy <= 0 || *total <= *buy {
		return qty
	}
	groups := qty / *total
	rem := qty % *total
	return groups**buy + rem
}

// ---------- handlers ----------

func (s *Server) categories(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(
		`SELECT slug, name, COALESCE(code,''), COALESCE(note,'')
		 FROM categories ORDER BY sort_order, id`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []Category{}
	for rows.Next() {
		var c Category
		_ = rows.Scan(&c.Slug, &c.Name, &c.Code, &c.Note)
		out = append(out, c)
	}
	writeJSON(w, out)
}

const productColumns = `p.id, p.slug, p.folder, p.name, p.variant,
	             p.price, p.market_price, p.discount_type, p.discount_percent,
	             p.bundle_buy_qty, p.bundle_total_qty, p.source,
	             c.slug, c.name`

func scanProduct(row interface{ Scan(...any) error }, p *Product) error {
	return row.Scan(&p.ID, &p.Slug, &p.Folder, &p.Name, &p.Variant,
		&p.Price, &p.MarketPrice, &p.DiscountType, &p.DiscountPercent,
		&p.BundleBuyQty, &p.BundleTotalQty, &p.Source,
		&p.CategorySlug, &p.CategoryName)
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request) {
	catSlug := r.URL.Query().Get("category")

	q := `SELECT ` + productColumns + `
	      FROM products p
	      JOIN categories c ON c.id = p.category_id
	      WHERE p.is_active = 1`
	args := []any{}
	if catSlug != "" {
		q += ` AND c.slug = ?`
		args = append(args, catSlug)
	}
	q += ` ORDER BY c.sort_order, p.sort_order, p.id`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []Product{}
	for rows.Next() {
		var p Product
		if err := scanProduct(rows, &p); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		p.FinalPrice = finalPrice(p.Price, p.DiscountType, p.DiscountPercent)
		p.Images = []string{}

		// картинки тянем сразу — но для каталога нужна только первая
		imgRows, err := s.DB.Query(
			`SELECT path FROM product_images
			 WHERE product_id = ? ORDER BY sort_order, id LIMIT 1`, p.ID)
		if err == nil {
			if imgRows.Next() {
				var path string
				_ = imgRows.Scan(&path)
				p.Images = append(p.Images, path)
			}
			imgRows.Close()
		}

		out = append(out, p)
	}
	writeJSON(w, out)
}

func (s *Server) product(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	var p Product
	err := scanProduct(s.DB.QueryRow(
		`SELECT `+productColumns+`
		 FROM products p
		 JOIN categories c ON c.id = p.category_id
		 WHERE p.slug = ? AND p.is_active = 1`, slug), &p)

	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	p.FinalPrice = finalPrice(p.Price, p.DiscountType, p.DiscountPercent)
	p.Images = []string{}

	imgRows, err := s.DB.Query(
		`SELECT path FROM product_images
		 WHERE product_id = ? ORDER BY sort_order, id`, p.ID)
	if err == nil {
		defer imgRows.Close()
		for imgRows.Next() {
			var path string
			_ = imgRows.Scan(&path)
			p.Images = append(p.Images, path)
		}
	}

	writeJSON(w, p)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
