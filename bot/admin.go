package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// AdminBot принимает /start <секрет> от менеджеров, запоминает их chat_id
// и рассылает уведомления о новых заказах (с квитанцией и юзернеймом
// партнёра, который привёл клиента) и о сообщениях с формы обратной связи.
type AdminBot struct {
	tg     *Client
	db     *sql.DB
	secret string
}

// StartAdminBot запускает бота в фоне и возвращает управляющий объект для
// отправки уведомлений из HTTP-хендлеров. Если token пустой, возвращает
// nil — все методы AdminBot безопасны для вызова на nil-получателе.
func StartAdminBot(db *sql.DB, token, secret string) *AdminBot {
	if token == "" {
		log.Println("ADMIN_BOT_TOKEN не задан — уведомления администраторам в Telegram отключены")
		return nil
	}
	ab := &AdminBot{tg: New(token), db: db, secret: secret}
	go func() {
		log.Println("админ-бот запущен")
		ab.tg.Run("adminbot", ab.handle)
	}()
	return ab
}

func (ab *AdminBot) handle(u Update) {
	if u.Message == nil {
		return
	}
	text := strings.TrimSpace(u.Message.Text)
	if !strings.HasPrefix(text, "/start") {
		return
	}

	parts := strings.Fields(text)
	if ab.secret == "" || len(parts) < 2 || parts[1] != ab.secret {
		_ = ab.tg.SendMessage(u.Message.Chat.ID,
			"Доступ закрыт. Отправьте /start <код доступа> — код задан в .env как ADMIN_BOT_SECRET.", nil)
		return
	}

	_, err := ab.db.Exec(
		`INSERT INTO admin_recipients (chat_id, username) VALUES (?,?)
		 ON DUPLICATE KEY UPDATE username = VALUES(username)`,
		u.Message.Chat.ID, u.Message.From.Username,
	)
	if err != nil {
		log.Println("adminbot: не удалось сохранить получателя:", err)
		return
	}
	_ = ab.tg.SendMessage(u.Message.Chat.ID,
		"Готово! Сюда будут приходить новые заказы, квитанции об оплате и сообщения из формы обратной связи.", nil)
}

func (ab *AdminBot) recipients() []int64 {
	rows, err := ab.db.Query(`SELECT chat_id FROM admin_recipients`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

// NotifyOrder уведомляет всех подключённых администраторов о заказе,
// готовом к проверке: сумма, состав, контакты покупателя, юзернейм
// партнёра (если заказ пришёл по реферальной ссылке) и квитанция об оплате.
func (ab *AdminBot) NotifyOrder(orderID int, customerName, phone, address string, total int, itemsText, referralUsername, receiptPath string) {
	if ab == nil {
		return
	}
	var refLine string
	if referralUsername != "" {
		refLine = "\nПривёл партнёр: @" + referralUsername
	}
	var addrLine string
	if address != "" {
		addrLine = "\nАдрес: " + address
	}
	text := fmt.Sprintf(
		"🛒 Заказ №%d ожидает подтверждения оплаты\n\nПокупатель: %s\nТелефон: %s%s%s\n\n%s\nИтого: %s ₽",
		orderID, customerName, phone, addrLine, refLine, itemsText, FmtRUB(total),
	)
	for _, chatID := range ab.recipients() {
		if err := ab.tg.SendMessage(chatID, text, nil); err != nil {
			log.Println("adminbot: sendMessage:", err)
		}
		if receiptPath != "" {
			if err := ab.tg.SendDocument(chatID, receiptPath, fmt.Sprintf("Квитанция к заказу №%d", orderID)); err != nil {
				log.Println("adminbot: sendDocument:", err)
			}
		}
	}
}

// NotifyContact уведомляет администраторов о новом сообщении с формы
// обратной связи на странице «Контакты».
func (ab *AdminBot) NotifyContact(name, phone, message string) {
	if ab == nil {
		return
	}
	text := fmt.Sprintf("✉️ Новое сообщение с формы обратной связи\n\nИмя: %s\nТелефон: %s\nСообщение: %s", name, phone, message)
	for _, chatID := range ab.recipients() {
		if err := ab.tg.SendMessage(chatID, text, nil); err != nil {
			log.Println("adminbot: sendMessage:", err)
		}
	}
}
