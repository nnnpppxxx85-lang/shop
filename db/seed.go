package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

type seedCategory struct {
	slug, name, code, note string
	sortOrder              int
}

type seedProduct struct {
	categorySlug, slug, folder, name, variant string
	price, marketPrice                        int
	discountType                              string
	discountPercent                           int
	bundleBuy, bundleTotal                    int
}

// SeedDemoData наполняет пустую базу несколькими товарами на основе
// папок с фото, которые уже лежат в templates/assets — чтобы после
// первого запуска сайт сразу было на чём смотреть. Если товары уже
// есть (админ начал вести каталог сам), сидер ничего не делает.
func SeedDemoData(conn *sql.DB) error {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []seedCategory{
		{"iphone", "iPhone", "01", "Смартфоны Apple", 1},
		{"macbook", "MacBook", "02", "Ноутбуки Apple", 2},
		{"airpods", "AirPods", "04", "Наушники Apple", 3},
		{"aksessuary-apple", "Аксессуары Apple", "08", "Клавиатуры, мыши и другое", 4},
	}
	for _, c := range categories {
		if _, err := conn.Exec(
			`INSERT IGNORE INTO categories (slug, name, code, note, sort_order) VALUES (?,?,?,?,?)`,
			c.slug, c.name, c.code, c.note, c.sortOrder,
		); err != nil {
			return err
		}
	}

	products := []seedProduct{
		{
			categorySlug: "iphone", slug: "apple-iphone-17e-black", folder: "Apple_iPhone_17e_black",
			name: "Apple iPhone 17e", variant: "256 ГБ · Black",
			price: 58400, marketPrice: 59500,
			discountType: "percent", discountPercent: 15,
		},
		{
			categorySlug: "macbook", slug: "apple-macbook-pro-16-m5-pro-silver-2026", folder: "Apple_MacBook_Pro_16_M5_Pro_silver_2026",
			name: "Apple MacBook Pro 16 M5 Pro", variant: "Silver · 2026",
			price: 289990, marketPrice: 0,
			discountType: "none",
		},
		{
			categorySlug: "airpods", slug: "airpods-pro-4", folder: "airoids_pro_4",
			name: "AirPods Pro 4", variant: "",
			price: 22990, marketPrice: 0,
			discountType: "bundle", bundleBuy: 2, bundleTotal: 3,
		},
		{
			categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-touch-id-numpad", folder: "Apple_Magic_NEW_Touch_ID_Num._keypad",
			name: "Apple Magic Keyboard", variant: "Touch ID + цифровой блок",
			price: 16990, marketPrice: 18990,
			discountType: "percent", discountPercent: 10,
		},
	}

	for i, p := range products {
		var catID int
		if err := conn.QueryRow(`SELECT id FROM categories WHERE slug = ?`, p.categorySlug).Scan(&catID); err != nil {
			log.Printf("seed: категория %s не найдена, пропускаю %s: %v", p.categorySlug, p.slug, err)
			continue
		}

		var marketPrice any
		if p.marketPrice > 0 {
			marketPrice = p.marketPrice
		}
		var bundleBuy, bundleTotal any
		if p.discountType == "bundle" {
			bundleBuy, bundleTotal = p.bundleBuy, p.bundleTotal
		}
		var variant any
		if p.variant != "" {
			variant = p.variant
		}

		res, err := conn.Exec(
			`INSERT INTO products
				(category_id, slug, folder, name, variant, price, market_price,
				 discount_type, discount_percent, bundle_buy_qty, bundle_total_qty,
				 source, is_active, sort_order)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,1,?)`,
			catID, p.slug, p.folder, p.name, variant, p.price, marketPrice,
			p.discountType, p.discountPercent, bundleBuy, bundleTotal,
			"demo-seed", i,
		)
		if err != nil {
			log.Printf("seed: не удалось добавить %s: %v", p.slug, err)
			continue
		}
		productID, _ := res.LastInsertId()

		images, err := ScanFolderImages(p.folder)
		if err != nil {
			log.Printf("seed: нет фото для %s (%s): %v", p.slug, p.folder, err)
			continue
		}
		for j, img := range images {
			_, _ = conn.Exec(
				`INSERT INTO product_images (product_id, path, sort_order) VALUES (?,?,?)`,
				productID, img, j,
			)
		}
	}

	log.Println("seed: добавлены демонстрационные товары из templates/assets")
	return nil
}

// ScanFolderImages возвращает пути к картинкам в templates/assets/<folder>,
// отсортированные по номеру в имени файла. Используется как сидером, так
// и ручкой админки, чтобы обе стороны видели одинаковый список.
func ScanFolderImages(folder string) ([]string, error) {
	dir := filepath.Join("templates", "assets", folder)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	names = NaturalSortStrings(names)
	out := make([]string, 0, len(names))
	for _, n := range names {
		if isImageFile(n) {
			out = append(out, "/assets/"+folder+"/"+n)
		}
	}
	return out, nil
}

func isImageFile(name string) bool {
	ext := filepath.Ext(name)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".PNG", ".JPG", ".JPEG", ".WEBP":
		return true
	default:
		return false
	}
}
