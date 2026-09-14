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
	sortOrder                                 int
}

// SeedDemoData наполняет пустую базу стартовым каталогом на основе папок
// с фото, которые уже лежат в templates/assets — чтобы после первого
// запуска сайт сразу было на чём смотреть. Если товары уже есть (админ
// начал вести каталог сам), сидер ничего не делает.
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
		{"imac", "iMac", "03", "Моноблоки Apple", 3},
		{"airpods", "AirPods", "04", "Наушники Apple", 4},
		{"apple-watch", "Apple Watch", "05", "Умные часы Apple", 5},
		{"monitory", "Мониторы", "06", "Apple Studio Display", 6},
		{"aksessuary", "Аксессуары", "07", "Зарядки, кабели и адаптеры", 7},
		{"aksessuary-apple", "Аксессуары Apple", "08", "Magic Keyboard, Pencil и другое", 8},
		{"dyson", "Dyson", "09", "Техника для ухода", 9},
		{"consoles", "Игровые консоли", "10", "PlayStation", 10},
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
		{categorySlug: "iphone", slug: "iphone-14-128-gb-blue", folder: "iphone_14_blue", name: "iPhone 14", variant: "128 ГБ · Blue", price: 38990, marketPrice: 39900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "iphone", slug: "iphone-14-128-gb-pink", folder: "iphone_14_pink", name: "iPhone 14", variant: "128 ГБ · Pink", price: 38990, marketPrice: 39900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "iphone", slug: "iphone-14-128-gb-white", folder: "iphone_14_white", name: "iPhone 14", variant: "128 ГБ · White", price: 38990, marketPrice: 39900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "iphone", slug: "iphone-18-pro-max-256-gb", folder: "iphone_18_pro_max", name: "iPhone 18 Pro Max", variant: "256 ГБ", price: 174990, marketPrice: 179900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "iphone", slug: "apple-iphone-17e-256-gb-black", folder: "Apple_iPhone_17e_black", name: "Apple iPhone 17e", variant: "256 ГБ · Black", price: 58400, marketPrice: 59500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "iphone", slug: "apple-iphone-17e-256-gb-pink", folder: "Apple_iPhone_17e_pink", name: "Apple iPhone 17e", variant: "256 ГБ · Pink", price: 58400, marketPrice: 59500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "iphone", slug: "apple-iphone-17e-256-gb-white", folder: "Apple_iPhone_17e_white", name: "Apple iPhone 17e", variant: "256 ГБ · White", price: 58400, marketPrice: 59500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "macbook", slug: "macbook-air-13-m5-midnight", folder: "MacBook_Air_13_M5_Midnigh", name: "MacBook Air 13″ M5", variant: "Midnight", price: 126990, marketPrice: 129990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "macbook", slug: "macbook-air-13-m5-starlight", folder: "MacBook_Air_13_M5_Starlight", name: "MacBook Air 13″ M5", variant: "Starlight", price: 126990, marketPrice: 129990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "macbook", slug: "macbook-air-13-m5-sky-blue", folder: "MacBook_Air_13_M5_blue", name: "MacBook Air 13″ M5", variant: "Sky Blue", price: 126990, marketPrice: 129990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "macbook", slug: "macbook-air-13-m5-silver", folder: "MacBook_Air_13_M5_silver", name: "MacBook Air 13″ M5", variant: "Silver", price: 126990, marketPrice: 129990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "macbook", slug: "macbook-air-15-m5-midnight", folder: "Apple_MacBook_Air_15_M5_Midnight", name: "MacBook Air 15″ M5", variant: "Midnight", price: 155990, marketPrice: 159990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "macbook", slug: "macbook-air-15-m5-starlight", folder: "Apple_MacBook_Air_15_M5_Starlight", name: "MacBook Air 15″ M5", variant: "Starlight", price: 155990, marketPrice: 159990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "macbook", slug: "macbook-air-15-m5-silver", folder: "Apple_MacBook_Air_15_M5_silver", name: "MacBook Air 15″ M5", variant: "Silver", price: 155990, marketPrice: 159990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "macbook", slug: "macbook-air-15-m5-sky-blue", folder: "Apple_MacBook_Air_15_M5_sky_blue", name: "MacBook Air 15″ M5", variant: "Sky Blue", price: 155990, marketPrice: 159990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 7},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-silver", folder: "Apple_MacBook_Pro_14_M5_silver", name: "MacBook Pro 14″ M5", variant: "Silver", price: 184990, marketPrice: 189990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 8},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-space-black", folder: "Apple_MacBook_Pro_14_M5_space_black", name: "MacBook Pro 14″ M5", variant: "Space Black", price: 184990, marketPrice: 189990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 9},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-pro-silver-2026", folder: "Apple_MacBook_Pro_14_M5_Pro_silver_2026", name: "MacBook Pro 14″ M5 Pro", variant: "Silver · 2026", price: 253990, marketPrice: 259990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 10},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-pro-space-black-2026", folder: "Apple_MacBook_Pro_14_M5_Pro_space_black_2026", name: "MacBook Pro 14″ M5 Pro", variant: "Space Black · 2026", price: 253990, marketPrice: 259990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 11},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-max-silver-2026", folder: "Apple_MacBook_Pro_14_M5_Max_silver_2026", name: "MacBook Pro 14″ M5 Max", variant: "Silver · 2026", price: 414990, marketPrice: 421194, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 12},
		{categorySlug: "macbook", slug: "macbook-pro-14-m5-max-space-black-2026", folder: "Apple_MacBook_Pro_14_M5_Max_space_black_2026", name: "MacBook Pro 14″ M5 Max", variant: "Space Black · 2026", price: 414990, marketPrice: 421194, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 13},
		{categorySlug: "macbook", slug: "macbook-pro-16-m5-pro-space-black", folder: "MacBook_Pro_16_M5_Pro_space_black", name: "MacBook Pro 16″ M5 Pro", variant: "Space Black", price: 343990, marketPrice: 350000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 14},
		{categorySlug: "macbook", slug: "macbook-pro-16-m5-pro-silver-2026", folder: "Apple_MacBook_Pro_16_M5_Pro_silver_2026", name: "MacBook Pro 16″ M5 Pro", variant: "Silver · 2026", price: 343990, marketPrice: 350000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 15},
		{categorySlug: "macbook", slug: "macbook-pro-16-m5-max-silver", folder: "Apple_MacBook_Pro_16_M5_Max_silver", name: "MacBook Pro 16″ M5 Max", variant: "Silver", price: 492990, marketPrice: 500000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 16},
		{categorySlug: "macbook", slug: "macbook-pro-16-m5-max-space-black", folder: "Apple_MacBook_Pro_16_M5_Max_space_black", name: "MacBook Pro 16″ M5 Max", variant: "Space Black", price: 492990, marketPrice: 500000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 17},
		{categorySlug: "imac", slug: "imac-24-m4-blue", folder: "iMac_24_Apple_M4_Blue", name: "iMac 24″ M4", variant: "Blue", price: 149990, marketPrice: 154500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "imac", slug: "imac-24-m4-green", folder: "iMac_24_Apple_M4_Green", name: "iMac 24″ M4", variant: "Green", price: 149990, marketPrice: 154500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "imac", slug: "imac-24-m4-silver", folder: "iMac_24_Apple_M4_Silver", name: "iMac 24″ M4", variant: "Silver", price: 149990, marketPrice: 154500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "imac", slug: "imac-24-m4-pink", folder: "iMac_24_Apple_M4_pink", name: "iMac 24″ M4", variant: "Pink", price: 149990, marketPrice: 154500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "airpods", slug: "airpods-pro-3", folder: "airpods_pro_3", name: "AirPods Pro 3", variant: "", price: 22990, marketPrice: 23900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "airpods", slug: "airpods-pro-4", folder: "airoids_pro_4", name: "AirPods Pro 4", variant: "", price: 18990, marketPrice: 20000, discountType: "bundle", discountPercent: 0, bundleBuy: 2, bundleTotal: 3, sortOrder: 1},
		{categorySlug: "airpods", slug: "airpods-max-2-black", folder: "Apple_AirPods_Max_2_black", name: "AirPods Max 2", variant: "Black", price: 54900, marketPrice: 54900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "airpods", slug: "airpods-max-2-desert", folder: "Apple_AirPods_Max_2_desert", name: "AirPods Max 2", variant: "Desert", price: 54900, marketPrice: 54900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "airpods", slug: "airpods-max-2-pink", folder: "Apple_AirPods_Max_2_pink", name: "AirPods Max 2", variant: "Pink", price: 54900, marketPrice: 54900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "airpods", slug: "airpods-max-2-purple", folder: "Apple_AirPods_Max_2_purple", name: "AirPods Max 2", variant: "Purple", price: 54900, marketPrice: 54900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "airpods", slug: "airpods-max-2-silver", folder: "Apple_AirPods_Max_2_silver", name: "AirPods Max 2", variant: "Silver", price: 54900, marketPrice: 54900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "apple-watch", slug: "apple-watch-se-3-starlight-aluminium", folder: "Apple_Watch_SE_3_Starlight_Aluminium", name: "Apple Watch SE 3", variant: "Starlight Aluminium", price: 26990, marketPrice: 27900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "apple-watch", slug: "apple-watch-series-11-black", folder: "Apple_Watch_Series_11_black", name: "Apple Watch Series 11", variant: "Black", price: 31990, marketPrice: 32900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "apple-watch", slug: "apple-watch-series-11-grey", folder: "Apple_Watch_Series_11_grey", name: "Apple Watch Series 11", variant: "Grey", price: 31990, marketPrice: 32900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "apple-watch", slug: "apple-watch-ultra-3-black-titanium", folder: "Apple_Watch_Ultra_3_Black_Titanium", name: "Apple Watch Ultra 3", variant: "Black Titanium", price: 57990, marketPrice: 58990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "apple-watch", slug: "apple-watch-ultra-3-natural-titanium", folder: "Apple_Watch_Ultra_3_Natural_Titanium", name: "Apple Watch Ultra 3", variant: "Natural Titanium", price: 57990, marketPrice: 58990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "monitory", slug: "apple-studio-display-27-standard-glass", folder: "Apple_Studio_Display", name: "Apple Studio Display", variant: "27″ Standard glass", price: 164990, marketPrice: 170000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "monitory", slug: "apple-studio-display-xdr-27-2026", folder: "Apple_Studio_Display_XDR_27_Standard_glass_with_tilt-_and_height-adjustable_stand_2026", name: "Apple Studio Display XDR", variant: "27″ · 2026", price: 293990, marketPrice: 300000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-23-75k-gold-blue-gold", folder: "Dyson_HD07_Supersonic_23.75K_Gold_BlueGold", name: "Dyson Supersonic HD07", variant: "23.75K Gold / Blue-Gold", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-black-nickel", folder: "Dyson_HD07_Supersonic_BlackNickel", name: "Dyson Supersonic HD07", variant: "Black Nickel", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-ceramic-pink-rose-gold", folder: "Dyson_HD07_Supersonic_Ceramic_PinkRose_Gold", name: "Dyson Supersonic HD07", variant: "Ceramic Pink / Rose Gold", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-fuchsia", folder: "Dyson_HD07_Supersonic_Fuchsia", name: "Dyson Supersonic HD07", variant: "Fuchsia", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-gift-edition-topaz-orange", folder: "Dyson_HD07_Supersonic_Gift_Edition_Topaz_Orange", name: "Dyson Supersonic HD07", variant: "Gift Edition Topaz Orange", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-limited-edition-violet-blue-rose", folder: "Dyson_HD07_Supersonic_Limited_Edition_Violet_BlueRose", name: "Dyson Supersonic HD07", variant: "Limited Edition Violet / Blue-Rose", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-onyx-black", folder: "Dyson_HD07_Supersonic_Onyx_Black", name: "Dyson Supersonic HD07", variant: "Onyx Black", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-origin-nickel", folder: "Dyson_HD07_Supersonic_Origin_Nickel", name: "Dyson Supersonic HD07", variant: "Origin Nickel", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 7},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-prussian-blue-rich-copper", folder: "Dyson_HD07_Supersonic_Straight_Wavy_Prussian_BlueRich_Copper", name: "Dyson Supersonic HD07", variant: "Prussian Blue / Rich Copper", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 8},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd07-iron-fuchsia", folder: "Dyson_HD07_Supersonic_iron", name: "Dyson Supersonic HD07", variant: "Iron / Fuchsia", price: 37990, marketPrice: 38780, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 9},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd17-r-ceramic-pink-rose", folder: "Dyson_HD17_Supersonic_R_Ceramic_pinkRose", name: "Dyson Supersonic HD17 R", variant: "Ceramic Pink / Rose", price: 62990, marketPrice: 64900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 10},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd17-r-jasper-plum", folder: "Dyson_HD17_Supersonic_R_JasperPlum", name: "Dyson Supersonic HD17 R", variant: "Jasper Plum", price: 62990, marketPrice: 64900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 11},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd17-r-pro-ceramic-pink-rose", folder: "Dyson_HD17_Supersonic_R_Pro_Ceramic_PinkRose", name: "Dyson Supersonic HD17 R Pro", variant: "Ceramic Pink / Rose", price: 62990, marketPrice: 64900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 12},
		{categorySlug: "dyson", slug: "dyson-supersonic-hd17-r-red-velvet-gold-limited-edition", folder: "Dyson_HD17_Supersonic_R_Red_VelvetGold_Limitted_Edition", name: "Dyson Supersonic HD17 R", variant: "Red Velvet / Gold Limited Edition", price: 62990, marketPrice: 64900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 13},
		{categorySlug: "consoles", slug: "sony-playstation-5-slim-digital-edition-825gb", folder: "Sony_PlayStation_5_Slim_Digital_Edition_825GB", name: "Sony PlayStation 5 Slim", variant: "Digital Edition · 825GB", price: 54990, marketPrice: 55900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "consoles", slug: "sony-playstation-5-slim-1tb-s-diskovodom", folder: "Sony_PlayStation_5_Slim_1TB", name: "Sony PlayStation 5 Slim", variant: "1TB · с дисководом", price: 61990, marketPrice: 62900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "consoles", slug: "sony-playstation-5-pro-2tb", folder: "Sony_PlayStation_5_Pro_2TB", name: "Sony PlayStation 5 Pro", variant: "2TB", price: 87990, marketPrice: 89900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "consoles", slug: "sony-playstation-5-slim-30th-anniversary-limited-edition", folder: "Sony_PlayStation_5_Slim_Digital_Edition_30th_Anniversary_Limited_Edition_Bundle", name: "Sony PlayStation 5 Slim", variant: "30th Anniversary Limited Edition", price: 87990, marketPrice: 89900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-20w-usb-c", folder: "Apple_20W_USB-C_Power_Adapter", name: "Адаптер питания Apple 20W USB-C", variant: "", price: 2190, marketPrice: 2490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-30w-usb-c", folder: "Apple_30W_USB-C", name: "Адаптер питания Apple 30W USB-C", variant: "", price: 4990, marketPrice: 5490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-35w-dual-usb-c", folder: "Apple_35W_Dual_USB-C_Power_Adapter", name: "Адаптер питания Apple 35W Dual USB-C", variant: "", price: 6790, marketPrice: 7490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-40w-dynamic-do-60w-max", folder: "Apple_40W_Dynamic_Power_Adapter_with_60W_Max", name: "Адаптер питания Apple 40W Dynamic", variant: "до 60W Max", price: 6790, marketPrice: 7490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-70w-usb-c", folder: "Apple_70W_USB-C_Power_Adapter", name: "Адаптер питания Apple 70W USB-C", variant: "", price: 6790, marketPrice: 7490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-96w-usb-c", folder: "Apple_96W_USB-C_Power_Adapter", name: "Адаптер питания Apple 96W USB-C", variant: "", price: 7690, marketPrice: 8490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "aksessuary", slug: "adapter-pitaniya-apple-140w-usb-c", folder: "Apple_140W_USB-C_Power_Adapter", name: "Адаптер питания Apple 140W USB-C", variant: "", price: 9990, marketPrice: 10900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "aksessuary", slug: "apple-usb-c-digital-av-multiport-adapter-4k", folder: "Apple_USB-C_to_Digital_AV_Multiport_Adapter_4K", name: "Apple USB-C Digital AV Multiport Adapter", variant: "4K", price: 6290, marketPrice: 6990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 7},
		{categorySlug: "aksessuary", slug: "apple-usb-c-to-usb-adapter-mj1m2", folder: "Apple_USB-C_to_USB_Adapter_MJ1M2", name: "Apple USB-C to USB Adapter", variant: "MJ1M2", price: 3090, marketPrice: 3490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 8},
		{categorySlug: "aksessuary", slug: "apple-magsafe-battery-dlya-iphone-air", folder: "Apple_MagSafe_Battery_для_iPhone_Air", name: "Apple MagSafe Battery", variant: "для iPhone Air", price: 67990, marketPrice: 70000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 9},
		{categorySlug: "aksessuary", slug: "kabel-usb-c-charge-cable-1-m", folder: "USB-C_Charge_Cable", name: "Кабель USB-C Charge Cable", variant: "1 м", price: 2590, marketPrice: 2990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 10},
		{categorySlug: "aksessuary", slug: "kabel-usb-c-charge-cable-2-m", folder: "USB-C_Charge_Cable_2_м", name: "Кабель USB-C Charge Cable", variant: "2 м", price: 3990, marketPrice: 4490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 11},
		{categorySlug: "aksessuary", slug: "kabel-usb-c-to-lightning-1-m", folder: "USB-C_to_Lightning_Cable", name: "Кабель USB-C to Lightning", variant: "1 м", price: 2990, marketPrice: 3490, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 12},
		{categorySlug: "aksessuary", slug: "kabel-usb-c-to-magsafe-3-2-m", folder: "USB-C_to_MagSafe_3_Cable", name: "Кабель USB-C to MagSafe 3", variant: "2 м", price: 6290, marketPrice: 6990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 13},
		{categorySlug: "aksessuary", slug: "kabel-usb-c-magsafe-dlya-apple-watch-fast-charger", folder: "USB-C_to_Apple_Watch_Magnetic_Fast_Charger", name: "Кабель USB-C — MagSafe для Apple Watch", variant: "Fast Charger", price: 5290, marketPrice: 5990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 14},
		{categorySlug: "aksessuary", slug: "kabel-thunderbolt-4-pro", folder: "Thunderbolt_4_Pro_Cable", name: "Кабель Thunderbolt 4 Pro", variant: "", price: 8990, marketPrice: 9990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 15},
		{categorySlug: "aksessuary", slug: "kabel-thunderbolt-5-pro", folder: "Thunderbolt_5_Pro_Cable", name: "Кабель Thunderbolt 5 Pro", variant: "", price: 11990, marketPrice: 12990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 16},
		{categorySlug: "aksessuary", slug: "kronshtein-dlya-monitora-gembird-ma-da2-06-17-32-black", folder: "Gembird_MA-DA2-06_17-32_Black", name: "Кронштейн для монитора Gembird MA-DA2-06", variant: "17-32″ · Black", price: 3590, marketPrice: 3990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 17},
		{categorySlug: "aksessuary", slug: "kronshtein-dlya-monitora-gembird-ma-da2-06-17-32-white", folder: "Gembird_MA-DA2-06-W_17-32_White", name: "Кронштейн для монитора Gembird MA-DA2-06", variant: "17-32″ · White", price: 3590, marketPrice: 3990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 18},
		{categorySlug: "aksessuary", slug: "treker-apple-airtag", folder: "Трекер_Apple_AirTag", name: "Трекер Apple AirTag", variant: "", price: 3490, marketPrice: 3990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 19},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-touch-id-cifrovoi-blok", folder: "Apple_Magic_NEW_Touch_ID_Num._keypad", name: "Apple Magic Keyboard", variant: "Touch ID + цифровой блок", price: 13990, marketPrice: 15000, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 0},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-usb-c-white", folder: "Apple_Magic_Keyboard_USB-C_white", name: "Apple Magic Keyboard", variant: "USB-C · White", price: 9990, marketPrice: 10900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 1},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-white", folder: "Apple_Magic_Keyboard_White", name: "Apple Magic Keyboard", variant: "White", price: 12990, marketPrice: 14100, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 2},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-sovmestimaya-white", folder: "Apple_Magic_Keyboard_white_oem", name: "Apple Magic Keyboard (совместимая)", variant: "White", price: 12990, marketPrice: 14100, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 3},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-s-touch-id-usb-c-white", folder: "Magic_Keyboard_with_Touch_ID_USB-C_white", name: "Apple Magic Keyboard с Touch ID", variant: "USB-C · White", price: 15990, marketPrice: 16900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 4},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-mouse-black", folder: "Apple_Magic_Mouse_Multi-Touch_Surface_Black", name: "Apple Magic Mouse", variant: "Black", price: 9990, marketPrice: 10900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 5},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-mouse-white", folder: "Apple_Magic_Mouse_Multi-Touch_Surface_White", name: "Apple Magic Mouse", variant: "White", price: 9990, marketPrice: 10900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 6},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-trackpad-black", folder: "Apple_Magic_Trackpad_Black", name: "Apple Magic Trackpad", variant: "Black", price: 14990, marketPrice: 15900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 7},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-trackpad-white", folder: "Apple_Magic_Trackpad_White", name: "Apple Magic Trackpad", variant: "White", price: 14990, marketPrice: 15900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 8},
		{categorySlug: "aksessuary-apple", slug: "apple-pencil-2-go-pokoleniya", folder: "Apple_Pencil_2-го_поколения", name: "Apple Pencil", variant: "2-го поколения", price: 9990, marketPrice: 10900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 9},
		{categorySlug: "aksessuary-apple", slug: "apple-pencil-pro", folder: "Apple_Pencil_Pro", name: "Apple Pencil Pro", variant: "", price: 11490, marketPrice: 12500, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 10},
		{categorySlug: "aksessuary-apple", slug: "apple-pencil-usb-c", folder: "Apple_Pencil_USB-C", name: "Apple Pencil USB-C", variant: "", price: 7990, marketPrice: 8990, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 11},
		{categorySlug: "aksessuary-apple", slug: "apple-magic-keyboard-folio-dlya-ipad", folder: "Apple_Magic_Keyboard_Folio_для_iPad", name: "Apple Magic Keyboard Folio", variant: "для iPad", price: 28990, marketPrice: 29900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 12},
		{categorySlug: "aksessuary-apple", slug: "magic-keyboard-dlya-ipad-pro-11-white", folder: "Magic_Keyboard_для_iPad_Pro_11″_white", name: "Magic Keyboard для iPad Pro 11″", variant: "White", price: 30990, marketPrice: 31900, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 13},
		{categorySlug: "aksessuary-apple", slug: "magic-keyboard-dlya-ipad-pro-13-black", folder: "Magic_Keyboard_для_iPad_Pro_13″_black", name: "Magic Keyboard для iPad Pro 13″", variant: "Black", price: 41490, marketPrice: 42600, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 14},
		{categorySlug: "aksessuary-apple", slug: "magic-keyboard-dlya-ipad-pro-13-white", folder: "Magic_Keyboard_для_iPad_Pro_13″_white", name: "Magic Keyboard для iPad Pro 13″", variant: "White", price: 41490, marketPrice: 42600, discountType: "none", discountPercent: 0, bundleBuy: 0, bundleTotal: 0, sortOrder: 15},
	}

	for _, p := range products {
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
			nil, p.sortOrder,
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

	log.Printf("seed: добавлен стартовый каталог (%d товаров) из templates/assets", len(products))
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
