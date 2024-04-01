package models

// Order представляет данные о заявке
type Order struct {
	OrderUID          string                   `json:"order_uid"`
	TrackNumber       string                   `json:"track_number"`
	Entry             string                   `json:"entry"`
	Delivery          map[string]interface{}   `json:"delivery"`
	Payment           map[string]interface{}   `json:"payment"`
	Items             []map[string]interface{} `json:"items"`
	Locale            string                   `json:"locale"`
	InternalSignature string                   `json:"internal_signature"`
	CustomerID        string                   `json:"customer_id"`
	DeliveryService   string                   `json:"delivery_service"`
	Shardkey          string                   `json:"shardkey"`
	SMID              int                      `json:"sm_id"`
	DateCreated       string                   `json:"date_created"`
	OofShard          string                   `json:"oof_shard"`
}
