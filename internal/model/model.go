package model

// Delivery описывает информацию о доставке.
type Delivery struct {
	Name    *string `json:"name,omitempty"`    // Имя получателя
	Phone   *string `json:"phone,omitempty"`   // Телефонный номер получателя
	Zip     *string `json:"zip,omitempty"`     // Почтовый индекс
	City    *string `json:"city,omitempty"`    // Город
	Address *string `json:"address,omitempty"` // Адрес
	Region  *string `json:"region,omitempty"`  // Регион
	Email   *string `json:"email,omitempty"`   // Электронная почта
}

// Payment описывает информацию об оплате.
type Payment struct {
	Transaction  *string `json:"transaction,omitempty"`   // Идентификатор транзакции
	RequestID    *string `json:"request_id,omitempty"`    // Идентификатор запроса
	Currency     *string `json:"currency,omitempty"`      // Валюта
	Provider     *string `json:"provider,omitempty"`      // Провайдер платежа
	Amount       *int    `json:"amount,omitempty"`        // Сумма
	PaymentDt    *int    `json:"payment_dt,omitempty"`    // Дата и время платежа
	Bank         *string `json:"bank,omitempty"`          // Банк
	DeliveryCost *int    `json:"delivery_cost,omitempty"` // Стоимость доставки
	GoodsTotal   *int    `json:"goods_total,omitempty"`   // Общая стоимость товаров
	CustomFee    *int    `json:"custom_fee,omitempty"`    // Сборы
}

// Item описывает информацию о товаре в заказе.
type Item struct {
	ChrtID      *int    `json:"chrt_id,omitempty"`      // Идентификатор товара
	TrackNumber *string `json:"track_number,omitempty"` // Номер отслеживания
	Price       *int    `json:"price,omitempty"`        // Цена
	RID         *string `json:"rid,omitempty"`          // Внутренний идентификатор
	Name        *string `json:"name,omitempty"`         // Название
	Sale        *int    `json:"sale,omitempty"`         // Скидка
	Size        *string `json:"size,omitempty"`         // Размер
	TotalPrice  *int    `json:"total_price,omitempty"`  // Итоговая цена
	NmID        *int    `json:"nm_id,omitempty"`        // Внешний идентификатор
	Brand       *string `json:"brand,omitempty"`        // Бренд
	Status      *int    `json:"status,omitempty"`       // Статус
}

// Order описывает структуру заказа.
type Order struct {
	OrderUID          string    `json:"order_uid"`                    // Уникальный идентификатор заказа (обязательное поле)
	TrackNumber       *string   `json:"track_number,omitempty"`       // Номер отслеживания заказа
	Entry             *string   `json:"entry,omitempty"`              // Точка входа
	Delivery          *Delivery `json:"delivery,omitempty"`           // Информация о доставке (опционально)
	Payment           *Payment  `json:"payment,omitempty"`            // Информация об оплате (опционально)
	Items             []Item    `json:"items"`                        // Список товаров (может быть пустым, но не nil)
	Locale            *string   `json:"locale,omitempty"`             // Локализация
	InternalSignature *string   `json:"internal_signature,omitempty"` // Внутренняя подпись
	CustomerID        *string   `json:"customer_id,omitempty"`        // Идентификатор клиента
	DeliveryService   *string   `json:"delivery_service,omitempty"`   // Служба доставки
	Shardkey          *string   `json:"shardkey,omitempty"`           // Ключ шардирования
	SMID              *int      `json:"sm_id,omitempty"`              // Идентификатор социальных медиа
	DateCreated       string    `json:"date_created"`                 // Дата создания заказа (обязательное поле)
	OofShard          *string   `json:"oof_shard,omitempty"`          // Шард для OOF
}
