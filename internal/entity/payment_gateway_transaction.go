package entity

import "time"

// PaymentGatewayTransaction represents the payment_gateway_transactions table
type PaymentGatewayTransaction struct {
	ID           string                `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	InvoiceID    string                `gorm:"column:invoice_id;type:uuid;not null"`
	Gateway      string                `gorm:"column:gateway;type:varchar(50);not null"`
	OrderID      string                `gorm:"column:order_id;type:varchar(100);not null"`
	PaymentURL   string                `gorm:"column:payment_url;type:text"`
	Token        string                `gorm:"column:token;type:varchar(255)"`
	Amount       float64               `gorm:"column:amount;type:decimal(10,2);not null"`
	Status       PaymentGatewayStatus  `gorm:"column:status;type:payment_status;default:'pending'"`
	PaymentType  string                `gorm:"column:payment_type;type:varchar(50)"`
	FraudStatus  string                `gorm:"column:fraud_status;type:varchar(50)"`
	CreatedAt    time.Time             `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time             `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Invoice *Invoice `gorm:"foreignKey:InvoiceID"`
}

func (PaymentGatewayTransaction) TableName() string { return "payment_gateway_transactions" }