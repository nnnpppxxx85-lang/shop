// Package bot реализует двух Telegram-ботов магазина (реферальный и
// админский) поверх минимального самодельного клиента Bot API — без
// внешних зависимостей, только stdlib, как и весь остальной проект.
package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Client struct {
	token  string
	client *http.Client
}

func New(token string) *Client {
	return &Client{token: token, client: &http.Client{Timeout: 65 * time.Second}}
}

func (c *Client) apiURL(method string) string {
	return "https://api.telegram.org/bot" + c.token + "/" + method
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type Message struct {
	MessageID int64     `json:"message_id"`
	From      User      `json:"from"`
	Chat      Chat      `json:"chat"`
	Text      string    `json:"text"`
	Document  *Document `json:"document"`
}

type CallbackQuery struct {
	ID      string  `json:"id"`
	From    User    `json:"from"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type apiEnvelope struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

func (c *Client) call(method string, body any, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := c.client.Post(c.apiURL(method), "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return fmt.Errorf("telegram %s: %s", method, env.Description)
	}
	if out != nil && len(env.Result) > 0 {
		return json.Unmarshal(env.Result, out)
	}
	return nil
}

// GetUpdates делает long-polling запрос к Bot API (до 50 секунд ожидания).
func (c *Client) GetUpdates(offset int64) ([]Update, error) {
	q := url.Values{}
	q.Set("timeout", "50")
	q.Set("offset", strconv.FormatInt(offset, 10))
	resp, err := c.client.Get(c.apiURL("getUpdates") + "?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if !env.OK {
		return nil, fmt.Errorf("telegram getUpdates: %s", env.Description)
	}
	var updates []Update
	if err := json.Unmarshal(env.Result, &updates); err != nil {
		return nil, err
	}
	return updates, nil
}

func (c *Client) SendMessage(chatID int64, text string, markup *InlineKeyboardMarkup) error {
	body := map[string]any{"chat_id": chatID, "text": text}
	if markup != nil {
		body["reply_markup"] = markup
	}
	return c.call("sendMessage", body, nil)
}

func (c *Client) AnswerCallbackQuery(id string) error {
	return c.call("answerCallbackQuery", map[string]any{"callback_query_id": id}, nil)
}

// SendDocument отправляет файл с диска (например, PDF-квитанцию) с подписью.
func (c *Client) SendDocument(chatID int64, filePath, caption string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}
	part, err := w.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	resp, err := c.client.Post(c.apiURL("sendDocument"), w.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var env apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return fmt.Errorf("telegram sendDocument: %s", env.Description)
	}
	return nil
}

// Run запускает бесконечный long-polling цикл и передаёт каждое
// обновление в handler. Сетевые сбои не останавливают бота.
func (c *Client) Run(name string, handler func(Update)) {
	var offset int64
	for {
		updates, err := c.GetUpdates(offset)
		if err != nil {
			log.Printf("%s: getUpdates: %v", name, err)
			time.Sleep(3 * time.Second)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			handler(u)
		}
	}
}

// FmtRUB форматирует сумму в рублях с пробелами между разрядами тысяч.
func FmtRUB(n int) string {
	s := strconv.Itoa(n)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	var groups []string
	for len(s) > 3 {
		groups = append([]string{s[len(s)-3:]}, groups...)
		s = s[:len(s)-3]
	}
	groups = append([]string{s}, groups...)
	out := ""
	for i, g := range groups {
		if i > 0 {
			out += " "
		}
		out += g
	}
	if neg {
		out = "-" + out
	}
	return out
}
