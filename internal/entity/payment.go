package entity

import "time"


type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodCard     PaymentMethod = "card"
)

// Payment represents the payments table
type Payment struct {
	ID              string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	InvoiceID       string     `gorm:"column:invoice_id;type:uuid;not null"`
	Amount          float64    `gorm:"column:amount;type:decimal(10,2);not null"`
	PaymentDate     time.Time  `gorm:"column:payment_date;type:timestamptz;default:CURRENT_TIMESTAMP"`
	PaymentMethod   string     `gorm:"column:payment_method;type:varchar(50);not null"`
	ReferenceNumber string     `gorm:"column:reference_number;type:varchar(100)"`
	Notes           string     `gorm:"column:notes;type:text"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Invoice *Invoice `gorm:"foreignKey:InvoiceID"`
}

func (Payment) TableName() string { return "payments" }
