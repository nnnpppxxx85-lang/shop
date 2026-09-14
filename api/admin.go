package api

import (
	"dragon/db"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type productInput struct {
	CategorySlug    string   `json:"categorySlug"`
	Slug            string   `json:"slug"`
	Folder          string   `json:"folder"`
	Name            string   `json:"name"`
	Variant         string   `json:"variant"`
	Price           int      `json:"price"`
	MarketPrice     int      `json:"marketPrice"`
	DiscountType    string   `json:"discountType"`
	DiscountPercent int      `json:"discountPercent"`
	BundleBuyQty    int      `json:"bundleBuyQty"`
	BundleTotalQty  int      `json:"bundleTotalQty"`
	Source          string   `json:"source"`
	IsActive        *bool    `json:"isActive"`
	SortOrder       int      `json:"sortOrder"`
	Images          []string `json:"images"`
}

// validate приводит вход к согласованному виду и отклоняет то, что не
// удастся честно отобразить на сайте (например, скидку 99% или комплект
// «купи 3 — получи 2»).
func (in *productInput) validate() error {
	in.CategorySlug = strings.TrimSpace(in.CategorySlug)
	in.Slug = strings.TrimSpace(in.Slug)
	in.Folder = strings.TrimSpace(in.Folder)
	in.Name = strings.TrimSpace(in.Name)

	if in.CategorySlug == "" || in.Slug == "" || in.Name == "" {
		return errors.New("заполните категорию, название и slug")
	}
	if in.Price <= 0 {
		return errors.New("цена должна быть больше нуля")
	}
	if in.MarketPrice < 0 {
		return errors.New("цена рынка не может быть отрицательной")
	}

	switch in.DiscountType {
	case "", "none":
		in.DiscountType = "none"
		in.DiscountPercent, in.BundleBuyQty, in.BundleTotalQty = 0, 0, 0
	case "percent":
		if in.DiscountPercent <= 0 || in.DiscountPercent > 90 {
			return errors.New("скидка в процентах должна быть от 1 до 90")
		}
		in.BundleBuyQty, in.BundleTotalQty = 0, 0
	case "bundle":
		if in.BundleBuyQty < 1 || in.BundleTotalQty <= in.BundleBuyQty {
			return errors.New(`для комплекта укажите "купить" ≥ 1 и "всего" больше "купить" (например, 2 и 3 — это 2+1=3)`)
		}
		in.DiscountPercent = 0
	default:
		return errors.New("неизвестный тип скидки")
	}
	return nil
}

type adminProduct struct {
	Product
	IsActive  bool `json:"isActive"`
	SortOrder int  `json:"sortOrder"`
}

// GET /api/admin/products — все товары (в т.ч. неактивные)
func (s *Server) adminProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT ` + productColumns + `, p.is_active, p.sort_order
		FROM products p JOIN categories c ON c.id = p.category_id
		ORDER BY c.sort_order, p.sort_order, p.id`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []adminProduct{}
	for rows.Next() {
		var p adminProduct
		if err := rows.Scan(&p.ID, &p.Slug, &p.Folder, &p.Name, &p.Variant,
			&p.Price, &p.MarketPrice, &p.DiscountType, &p.DiscountPercent,
			&p.BundleBuyQty, &p.BundleTotalQty, &p.Source,
			&p.CategorySlug, &p.CategoryName, &p.IsActive, &p.SortOrder); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		p.FinalPrice = finalPrice(p.Price, p.DiscountType, p.DiscountPercent)
		p.Images = []string{}
		imgRows, err := s.DB.Query(`SELECT path FROM product_images WHERE product_id = ? ORDER BY sort_order, id`, p.ID)
		if err == nil {
			for imgRows.Next() {
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

func (s *Server) categoryID(slug string) (int, error) {
	var id int
	err := s.DB.QueryRow(`SELECT id FROM categories WHERE slug = ?`, slug).Scan(&id)
	return id, err
}

func (s *Server) saveImages(productID int64, images []string) {
	_, _ = s.DB.Exec(`DELETE FROM product_images WHERE product_id = ?`, productID)
	i := 0
	for _, img := range images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		_, _ = s.DB.Exec(`INSERT INTO product_images (product_id, path, sort_order) VALUES (?,?,?)`, productID, img, i)
		i++
	}
}

// POST /api/admin/products
func (s *Server) adminCreateProduct(w http.ResponseWriter, r *http.Request) {
	var in productInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	if err := in.validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	catID, err := s.categoryID(in.CategorySlug)
	if err != nil {
		http.Error(w, "категория не найдена", 400)
		return
	}
	active := 1
	if in.IsActive != nil && !*in.IsActive {
		active = 0
	}
	res, err := s.DB.Exec(`INSERT INTO products
		(category_id, slug, folder, name, variant, price, market_price,
		 discount_type, discount_percent, bundle_buy_qty, bundle_total_qty,
		 source, is_active, sort_order)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		catID, in.Slug, in.Folder, in.Name, nullStr(in.Variant), in.Price, nullInt(in.MarketPrice),
		in.DiscountType, in.DiscountPercent, nullInt(in.BundleBuyQty), nullInt(in.BundleTotalQty),
		nullStr(in.Source), active, in.SortOrder)
	if err != nil {
		writeDBError(w, err, "товар")
		return
	}
	id, _ := res.LastInsertId()
	s.saveImages(id, in.Images)
	writeJSON(w, map[string]any{"ok": true, "id": id})
}

// PUT /api/admin/products/{id}
func (s *Server) adminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	var in productInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	if err := in.validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	catID, err := s.categoryID(in.CategorySlug)
	if err != nil {
		http.Error(w, "категория не найдена", 400)
		return
	}
	active := 1
	if in.IsActive != nil && !*in.IsActive {
		active = 0
	}
	_, err = s.DB.Exec(`UPDATE products SET category_id=?, slug=?, folder=?, name=?, variant=?,
		price=?, market_price=?, discount_type=?, discount_percent=?, bundle_buy_qty=?, bundle_total_qty=?,
		source=?, is_active=?, sort_order=? WHERE id=?`,
		catID, in.Slug, in.Folder, in.Name, nullStr(in.Variant), in.Price, nullInt(in.MarketPrice),
		in.DiscountType, in.DiscountPercent, nullInt(in.BundleBuyQty), nullInt(in.BundleTotalQty),
		nullStr(in.Source), active, in.SortOrder, id)
	if err != nil {
		writeDBError(w, err, "товар")
		return
	}
	s.saveImages(int64(id), in.Images)
	writeJSON(w, map[string]any{"ok": true})
}

// DELETE /api/admin/products/{id}
func (s *Server) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if _, err := s.DB.Exec(`DELETE FROM product_images WHERE product_id = ?`, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if _, err := s.DB.Exec(`DELETE FROM products WHERE id = ?`, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// ---------- категории ----------

type categoryInput struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Note      string `json:"note"`
	SortOrder int    `json:"sortOrder"`
}

// POST /api/admin/categories
func (s *Server) adminCreateCategory(w http.ResponseWriter, r *http.Request) {
	var in categoryInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" || in.Name == "" {
		http.Error(w, "заполните slug и название категории", 400)
		return
	}
	if _, err := s.DB.Exec(`INSERT INTO categories (slug, name, code, note, sort_order) VALUES (?,?,?,?,?)`,
		in.Slug, in.Name, nullStr(in.Code), nullStr(in.Note), in.SortOrder); err != nil {
		writeDBError(w, err, "категорию")
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// PUT /api/admin/categories/{slug}
func (s *Server) adminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var in categoryInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверный запрос", 400)
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		http.Error(w, "укажите название категории", 400)
		return
	}
	newSlug := strings.TrimSpace(in.Slug)
	if newSlug == "" {
		newSlug = slug
	}
	if _, err := s.DB.Exec(`UPDATE categories SET slug=?, name=?, code=?, note=?, sort_order=? WHERE slug=?`,
		newSlug, in.Name, nullStr(in.Code), nullStr(in.Note), in.SortOrder, slug); err != nil {
		writeDBError(w, err, "категорию")
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// DELETE /api/admin/categories/{slug}
func (s *Server) adminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var count int
	_ = s.DB.QueryRow(`SELECT COUNT(*) FROM products p JOIN categories c ON c.id = p.category_id WHERE c.slug = ?`, slug).Scan(&count)
	if count > 0 {
		http.Error(w, "нельзя удалить категорию — в ней ещё есть товары", 400)
		return
	}
	if _, err := s.DB.Exec(`DELETE FROM categories WHERE slug = ?`, slug); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

// ---------- фото из templates/assets ----------

var folderNameRe = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё0-9_.\-]+$`)

// GET /api/admin/assets — список папок с фото, чтобы не искать их вручную
func (s *Server) adminListAssetFolders(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir("templates/assets")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	folders := []string{}
	for _, e := range entries {
		if e.IsDir() && e.Name() != "js" {
			folders = append(folders, e.Name())
		}
	}
	sort.Strings(folders)
	writeJSON(w, folders)
}

// GET /api/admin/assets/{folder} — фото внутри папки, уже в виде путей /assets/...
func (s *Server) adminFolderImages(w http.ResponseWriter, r *http.Request) {
	folder := r.PathValue("folder")
	if folder == "" || strings.Contains(folder, "..") || !folderNameRe.MatchString(folder) {
		http.Error(w, "некорректное имя папки", 400)
		return
	}
	images, err := db.ScanFolderImages(folder)
	if err != nil || len(images) == 0 {
		http.Error(w, "в папке templates/assets/"+folder+" не найдено фото", 404)
		return
	}
	writeJSON(w, images)
}

// ---------- заказы ----------

// GET /api/admin/orders
func (s *Server) adminOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT id, customer_name, phone, COALESCE(address,''), COALESCE(comment,''),
		total, payment_method, status, created_at FROM orders ORDER BY id DESC LIMIT 200`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var id, total int
		var name, phone, address, comment, method, status, created string
		_ = rows.Scan(&id, &name, &phone, &address, &comment, &total, &method, &status, &created)
		items := []map[string]any{}
		ir, err := s.DB.Query(`SELECT name, price, qty, line_total FROM order_items WHERE order_id = ?`, id)
		if err == nil {
			for ir.Next() {
				var n string
				var price, qty, lineTotal int
				_ = ir.Scan(&n, &price, &qty, &lineTotal)
				items = append(items, map[string]any{"name": n, "price": price, "qty": qty, "lineTotal": lineTotal})
			}
			ir.Close()
		}
		out = append(out, map[string]any{
			"id": id, "name": name, "phone": phone, "address": address, "comment": comment,
			"total": total, "paymentMethod": method, "status": status, "createdAt": created, "items": items,
		})
	}
	writeJSON(w, out)
}

// ---------- утилиты ----------

func writeDBError(w http.ResponseWriter, err error, what string) {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			http.Error(w, "Такой slug уже используется — выберите другой", 400)
			return
		case 1452:
			http.Error(w, "Указанная категория не найдена", 400)
			return
		}
	}
	http.Error(w, "Не удалось сохранить "+what+": "+err.Error(), 500)
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullInt(i int) any {
	if i <= 0 {
		return nil
	}
	return i
}
