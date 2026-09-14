package db

import "encoding/json"

// StorageOption — вариант объёма памяти устройства (например, "256 ГБ")
// с доплатой к базовой цене товара. Хранится в products.storage_options
// одной JSON-строкой, чтобы не заводить отдельную таблицу ради простого
// списка "подпись + доплата".
type StorageOption struct {
	Label      string `json:"label"`
	PriceDelta int    `json:"priceDelta"`
}

// EncodeStorageOptions сериализует варианты памяти в JSON для колонки
// products.storage_options; пустой список даёт NULL.
func EncodeStorageOptions(opts []StorageOption) any {
	if len(opts) == 0 {
		return nil
	}
	b, err := json.Marshal(opts)
	if err != nil {
		return nil
	}
	return string(b)
}

// DecodeStorageOptions разбирает JSON из колонки storage_options обратно
// в список вариантов; NULL, пустая строка или битый JSON дают пустой список.
func DecodeStorageOptions(raw *string) []StorageOption {
	if raw == nil || *raw == "" {
		return []StorageOption{}
	}
	var opts []StorageOption
	if err := json.Unmarshal([]byte(*raw), &opts); err != nil {
		return []StorageOption{}
	}
	return opts
}
