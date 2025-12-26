package entity




// QueueSettings represents the mikrotik_queue_settings table
type QueueSettings struct {
	ProfileID          string   `gorm:"primaryKey;type:uuid;column:profile_id"`
	QueueType          string   `gorm:"column:queue_type;type:varchar(50);default:'default'"`
	ParentQueue        string   `gorm:"column:parent_queue;type:varchar(50)"`
	Priority           string   `gorm:"column:priority;type:varchar(20);default:'8'"`
	BurstLimitUp       string   `gorm:"column:burst_limit_up;type:varchar(50)"`
	BurstLimitDown     string   `gorm:"column:burst_limit_down;type:varchar(50)"`
	BurstThresholdUp   string   `gorm:"column:burst_threshold_up;type:varchar(50)"`
	BurstThresholdDown string   `gorm:"column:burst_threshold_down;type:varchar(50)"`
	BurstTime          string   `gorm:"column:burst_time;type:varchar(20);default:'0s'"`
	LimitAtUp          string   `gorm:"column:limit_at_up;type:varchar(50)"`
	LimitAtDown        string   `gorm:"column:limit_at_down;type:varchar(50)"`
	MaxLimitUp         string   `gorm:"column:max_limit_up;type:varchar(50)"`
	MaxLimitDown       string   `gorm:"column:max_limit_down;type:varchar(50)"`
	PacketMarks        []string `gorm:"column:packet_marks;type:text[]"`

	// Relations
	Profile *MikrotikProfile `gorm:"foreignKey:ProfileID"`
}

func (QueueSettings) TableName() string { return "mikrotik_queue_settings" }