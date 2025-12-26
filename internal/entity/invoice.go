package entity

import "time"

// Invoice represents the invoices table
type Invoice struct {
	ID              string          `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID      string          `gorm:"column:mikrotik_id;type:uuid;not null"`
	CustomerID      string          `gorm:"column:customer_id;type:uuid;not null"`
	PackageID       string          `gorm:"column:package_id;type:uuid;not null"`
	InvoiceNumber   string          `gorm:"column:invoice_number;type:varchar(50);not null"`
	Amount          float64         `gorm:"column:amount;type:decimal(10,2);not null"`
	DueDate         time.Time       `gorm:"column:due_date;type:date;not null"`
	Status          InvoiceStatus   `gorm:"column:status;type:invoice_status;default:'unpaid'"`
	InvoiceType     InvoiceType     `gorm:"column:invoice_type;type:invoice_type;default:'monthly'"`
	PackageName     string          `gorm:"column:package_name;type:varchar(100)"`
	Description     string          `gorm:"column:description;type:text"`
	PaymentDate     *time.Time      `gorm:"column:payment_date;type:timestamptz"`
	PaymentMethod   string          `gorm:"column:payment_method;type:varchar(50)"`
	PaymentGateway  string          `gorm:"column:payment_gateway;type:varchar(50)"`
	PaymentToken    string          `gorm:"column:payment_token;type:varchar(255)"`
	PaymentURL      string          `gorm:"column:payment_url;type:text"`
	PaymentStatus   PaymentGatewayStatus `gorm:"column:payment_status;type:payment_status;default:'pending'"`
	Notes           string          `gorm:"column:notes;type:text"`
	CreatedAt       time.Time       `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik                 *Mikrotik                 `gorm:"foreignKey:MikrotikID"`
	Customer                 *Customer                 `gorm:"foreignKey:CustomerID"`
	Package                  *Package                  `gorm:"foreignKey:PackageID"`
	Payments                 []Payment                 `gorm:"foreignKey:InvoiceID"`
	GatewayTransactions      []PaymentGatewayTransaction `gorm:"foreignKey:InvoiceID"`
	CollectorPayments        []CollectorPayment       `gorm:"foreignKey:InvoiceID"`
	AgentMonthlyPayments     []AgentMonthlyPayment    `gorm:"foreignKey:InvoiceID"`
	AgentPayments            []AgentPayment           `gorm:"foreignKey:InvoiceID"`
}

func (Invoice) TableName() string { return "invoices" }