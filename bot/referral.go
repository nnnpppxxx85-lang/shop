package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// ReferralBot — бот для партнёров: показывает статистику по приведённым
// заказам и по кнопке выдаёт персональную реферальную ссылку на сайт.
type ReferralBot struct {
	tg      *Client
	db      *sql.DB
	siteURL string
}

// StartReferralBot запускает бота в фоне. Если token пустой, бот не
// запускается (сайт при этом продолжает работать как обычно).
func StartReferralBot(db *sql.DB, token, siteURL string) {
	if token == "" {
		log.Println("REF_BOT_TOKEN не задан — реферальный бот отключён")
		return
	}
	rb := &ReferralBot{tg: New(token), db: db, siteURL: strings.TrimRight(siteURL, "/")}
	go func() {
		log.Println("реферальный бот запущен")
		rb.tg.Run("refbot", rb.handle)
	}()
}

func (rb *ReferralBot) handle(u Update) {
	if u.CallbackQuery != nil {
		rb.handleCallback(u.CallbackQuery)
		return
	}
	if u.Message == nil {
		return
	}
	rb.upsertPartner(u.Message.From.ID, u.Message.From.Username)
	rb.sendStats(u.Message.Chat.ID, u.Message.From.ID)
}

func (rb *ReferralBot) handleCallback(cb *CallbackQuery) {
	_ = rb.tg.AnswerCallbackQuery(cb.ID)
	rb.upsertPartner(cb.From.ID, cb.From.Username)

	switch cb.Data {
	case "create_link":
		link := fmt.Sprintf("%s/?ref=%d", rb.siteURL, cb.From.ID)
		_ = rb.tg.SendMessage(cb.Message.Chat.ID,
			"Ваша реферальная ссылка:\n"+link+"\n\nОтправляйте её клиентам — все заказы по этой ссылке будут закреплены за вами.",
			nil)
	case "refresh":
		rb.sendStats(cb.Message.Chat.ID, cb.From.ID)
	}
}

func (rb *ReferralBot) upsertPartner(telegramID int64, username string) {
	_, _ = rb.db.Exec(
		`INSERT INTO partners (telegram_id, username) VALUES (?,?)
		 ON DUPLICATE KEY UPDATE username = VALUES(username)`,
		telegramID, username,
	)
}

func (rb *ReferralBot) sendStats(chatID, telegramID int64) {
	var orders, revenue int
	_ = rb.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(o.total),0)
		 FROM orders o JOIN partners p ON p.id = o.referral_partner_id
		 WHERE p.telegram_id = ?`, telegramID,
	).Scan(&orders, &revenue)

	text := fmt.Sprintf(
		"👋 Партнёрская программа DragonMobile\n\n"+
			"Приведено заказов: %d\n"+
			"Их общая сумма: %s ₽\n\n"+
			"Нажмите «Создать ссылку», чтобы получить свою реферальную ссылку на сайт.",
		orders, FmtRUB(revenue),
	)
	markup := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{{Text: "🔗 Создать ссылку", CallbackData: "create_link"}},
		{{Text: "🔄 Обновить статистику", CallbackData: "refresh"}},
	}}
	if err := rb.tg.SendMessage(chatID, text, markup); err != nil {
		log.Println("refbot: sendMessage:", err)
	}
}
