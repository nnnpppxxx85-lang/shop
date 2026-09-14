-- Оплата переводом (вместо СБП/QR), реферальная программа и уведомления
-- в Telegram. Применяется автоматически при старте сервера (см.
-- db/migrate.go), этот файл — история схемы для ручного применения.

ALTER TABLE orders
  ADD COLUMN referral_partner_id INT UNSIGNED DEFAULT NULL AFTER status,
  ADD COLUMN referral_username VARCHAR(191) DEFAULT NULL AFTER referral_partner_id,
  ADD COLUMN receipt_path VARCHAR(255) DEFAULT NULL AFTER referral_username;

CREATE TABLE IF NOT EXISTS settings (
  setting_key VARCHAR(64) PRIMARY KEY,
  value VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS partners (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  telegram_id BIGINT NOT NULL UNIQUE,
  username VARCHAR(191) DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_recipients (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  chat_id BIGINT NOT NULL UNIQUE,
  username VARCHAR(191) DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS contact_messages (
  id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(191) DEFAULT NULL,
  phone VARCHAR(64) DEFAULT NULL,
  message TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO settings (setting_key, value) VALUES
  ('payment_card', '2200 0000 0000 0000'),
  ('payment_phone', '+7 900 000-00-00'),
  ('consultant_telegram', 'https://t.me/dragonmobile_support');
