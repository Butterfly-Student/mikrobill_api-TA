package entity

import "time"


type CollectorPaymentStatus string


const (
	CollectorPaymentStatusCompleted CollectorPaymentStatus = "completed"
	CollectorPaymentStatusPending   CollectorPaymentStatus = "pending"
	CollectorPaymentStatusCancelled CollectorPaymentStatus = "cancelled"
)

// CollectorPayment represents the collector_payments table
type CollectorPayment struct {
	ID               string                `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CollectorID      string                `gorm:"column:collector_id;type:uuid;not null"`
	CustomerID       string                `gorm:"column:customer_id;type:uuid;not null"`
	InvoiceID        string                `gorm:"column:invoice_id;type:uuid;not null"`
	PaymentAmount    float64               `gorm:"column:payment_amount;type:decimal(15,2);not null"`
	CommissionAmount float64               `gorm:"column:commission_amount;type:decimal(15,2);not null"`
	PaymentMethod    PaymentMethod         `gorm:"column:payment_method;type:payment_method;default:'cash'"`
	PaymentDate      time.Time             `gorm:"column:payment_date;type:timestamptz;default:CURRENT_TIMESTAMP"`
	Notes            string                `gorm:"column:notes;type:text"`
	Status           CollectorPaymentStatus `gorm:"column:status;type:collector_payment_status;default:'completed'"`
	CreatedAt        time.Time             `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time             `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Collector *Collector `gorm:"foreignKey:CollectorID"`
	Customer  *Customer  `gorm:"foreignKey:CustomerID"`
	Invoice   *Invoice   `gorm:"foreignKey:InvoiceID"`
}

func (CollectorPayment) TableName() string { return "collector_payments" }