package entity

import "time"

type Expense struct {
	ID            string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Description   string     `gorm:"column:description;type:text;not null"`
	Amount        float64    `gorm:"column:amount;type:decimal(12,2);not null"`
	Category      string     `gorm:"column:category;type:varchar(100);not null"`
	ExpenseDate   time.Time  `gorm:"column:expense_date;type:date;not null"`
	PaymentMethod string     `gorm:"column:payment_method;type:varchar(50)"`
	Notes         string     `gorm:"column:notes;type:text"`
	CreatedBy     *string    `gorm:"column:created_by;type:uuid"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	CreatedByUser *User `gorm:"foreignKey:CreatedBy"`
}

func (Expense) TableName() string { return "expenses" }