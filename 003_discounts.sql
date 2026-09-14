-- Типы скидок на товары: обычный процент или "комплект" (напр. 2+1=3).
-- Применяется автоматически при старте сервера (см. db/migrate.go),
-- этот файл — история схемы для ручного применения при необходимости.

ALTER TABLE products
  ADD COLUMN discount_type ENUM('none','percent','bundle') NOT NULL DEFAULT 'none' AFTER discount_percent,
  ADD COLUMN bundle_buy_qty TINYINT UNSIGNED DEFAULT NULL AFTER discount_type,
  ADD COLUMN bundle_total_qty TINYINT UNSIGNED DEFAULT NULL AFTER bundle_buy_qty;

UPDATE products SET discount_type = 'percent' WHERE discount_percent > 0 AND discount_type = 'none';

ALTER TABLE order_items
  ADD COLUMN line_total INT UNSIGNED NOT NULL DEFAULT 0;

UPDATE order_items SET line_total = price * qty WHERE line_total = 0;
