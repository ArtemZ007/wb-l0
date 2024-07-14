package model

import (
	"github.com/go-playground/validator/v10"
)

// Delivery описывает информацию о доставке.
type Delivery struct {
	ID      *int    `json:"-" validate:"required"`                     // Идентификатор
	Name    *string `json:"name,omitempty" validate:"required"`        // Имя получателя
	Phone   *string `json:"phone,omitempty" validate:"required,e164"`  // Телефонный номер получателя
	Zip     *string `json:"zip,omitempty" validate:"required"`         // Почтовый индекс
	City    *string `json:"city,omitempty" validate:"required"`        // Город
	Address *string `json:"address,omitempty" validate:"required"`     // Адрес
	Region  *string `json:"region,omitempty" validate:"required"`      // Регион
	Email   *string `json:"email,omitempty" validate:"required,email"` // Электронная почта
}

// Payment описывает информацию об оплате.
type Payment struct {
	ID           *int    `json:"-" validate:"required"`                         // Идентификатор
	Transaction  *string `json:"transaction,omitempty" validate:"required"`     // Идентификатор транзакции
	RequestID    *string `json:"request_id,omitempty" validate:"required"`      // Идентификатор запроса
	Currency     *string `json:"currency,omitempty" validate:"required"`        // Валюта
	Provider     *string `json:"provider,omitempty" validate:"required"`        // Провайдер платежа
	Amount       *int    `json:"amount,omitempty" validate:"required,gt=0"`     // Сумма
	PaymentDt    *int64  `json:"payment_dt,omitempty" validate:"required,gt=0"` // Дата и время платежа
	Bank         *string `json:"bank,omitempty" validate:"required"`            // Банк
	DeliveryCost *int    `json:"delivery_cost,omitempty" validate:"gte=0"`      // Стоимость доставки
	GoodsTotal   *int    `json:"goods_total,omitempty" validate:"gte=0"`        // Общая стоимость товаров
	CustomFee    *int    `json:"custom_fee,omitempty" validate:"gte=0"`         // Сборы
}

// Item описывает информацию о товаре в заказе.
type Item struct {
	ID          *int    `json:"-" validate:"required"`                          // Идентификатор
	ChrtID      *int    `json:"chrt_id,omitempty" validate:"required"`          // Идентификатор товара
	TrackNumber *string `json:"track_number,omitempty" validate:"required"`     // Номер отслеживания
	Price       *int    `json:"price,omitempty" validate:"required,gt=0"`       // Цена
	RID         *string `json:"rid,omitempty" validate:"required"`              // Внутренний идентификатор
	Name        *string `json:"name,omitempty" validate:"required"`             // Название
	Sale        *int    `json:"sale,omitempty" validate:"gte=0,lte=100"`        // Скидка
	Size        *string `json:"size,omitempty" validate:"required"`             // Размер
	TotalPrice  *int    `json:"total_price,omitempty" validate:"required,gt=0"` // Итоговая цена
	NmID        *int    `json:"nm_id,omitempty" validate:"required"`            // Внешний идентификатор
	Brand       *string `json:"brand,omitempty" validate:"required"`            // Бренд
	Status      *int    `json:"status,omitempty" validate:"required"`           // Статус
}

// Order описывает структуру заказа.
type Order struct {
	ID                string    `json:"id" validate:"required,uuid4"`
	OrderUID          string    `json:"order_uid" validate:"required,uuid4"`               // Уникальный идентификатор заказа
	TrackNumber       *string   `json:"track_number,omitempty" validate:"omitempty,uuid4"` // Номер отслеживания заказа
	Entry             *string   `json:"entry,omitempty" validate:"omitempty"`              // Точка входа
	Delivery          *Delivery `json:"delivery,omitempty" validate:"omitempty,dive"`      // Информация о доставке
	Payment           *Payment  `json:"payment,omitempty" validate:"omitempty,dive"`       // Информация об оплате
	Items             []Item    `json:"items" validate:"required,dive"`                    // Список товаров
	Locale            *string   `json:"locale,omitempty" validate:"omitempty"`             // Локализация
	InternalSignature *string   `json:"internal_signature,omitempty" validate:"omitempty"` // Внутренняя подпись
	CustomerID        *string   `json:"customer_id,omitempty" validate:"omitempty,uuid4"`  // Идентификатор клиента
	DeliveryService   *string   `json:"delivery_service,omitempty" validate:"omitempty"`   // Служба доставки
	Shardkey          *string   `json:"shardkey,omitempty" validate:"omitempty"`           // Ключ шардирования
	SMID              *int      `json:"sm_id,omitempty" validate:"omitempty"`              // Идентификатор социальных медиа
	DateCreated       string    `json:"date_created" validate:"required"`                  // Дата создания заказа
	OofShard          *string   `json:"oof_shard,omitempty" validate:"omitempty"`          // Шард для OOF
}

// Validate выполняет валидацию структуры с использованием библиотеки go-playground/validator.
func (o *Order) Validate() error {
	validate := validator.New()
	return validate.Struct(o)
}
