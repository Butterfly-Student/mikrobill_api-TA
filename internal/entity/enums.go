package entity

// ENUM Types








type ConnectionType string
type CableCapacity string
type AgentStatus string
type AgentTransactionType string
type TransactionStatus string
type VoucherStatus string
type MaintenanceType string
type DeviceStatus string
type InvoiceStatus string
type InvoiceType string
type PaymentGatewayStatus string
type SegmentType string
type SegmentStatus string
type RequestStatus string
type NotificationType string

// Enum Constants
const (
	


	ConnectionTypeFiber     ConnectionType = "fiber"
	ConnectionTypeCopper    ConnectionType = "copper"
	ConnectionTypeWireless  ConnectionType = "wireless"
	ConnectionTypeMicrowave ConnectionType = "microwave"

	CableCapacity100M CableCapacity = "100M"
	CableCapacity1G   CableCapacity = "1G"
	CableCapacity10G  CableCapacity = "10G"
	CableCapacity100G CableCapacity = "100G"

	AgentStatusActive    AgentStatus = "active"
	AgentStatusInactive  AgentStatus = "inactive"
	AgentStatusSuspended AgentStatus = "suspended"

	AgentTransactionTypeDeposit        AgentTransactionType = "deposit"
	AgentTransactionTypeWithdrawal     AgentTransactionType = "withdrawal"
	AgentTransactionTypeVoucherSale    AgentTransactionType = "voucher_sale"
	AgentTransactionTypeMonthlyPayment AgentTransactionType = "monthly_payment"
	AgentTransactionTypeCommission     AgentTransactionType = "commission"
	AgentTransactionTypeBalanceRequest AgentTransactionType = "balance_request"

	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"

	VoucherStatusActive   VoucherStatus = "active"
	VoucherStatusUsed     VoucherStatus = "used"
	VoucherStatusExpired  VoucherStatus = "expired"
	VoucherStatusCancelled VoucherStatus = "cancelled"

	MaintenanceTypeRepair      MaintenanceType = "repair"
	MaintenanceTypeReplacement MaintenanceType = "replacement"
	MaintenanceTypeInspection  MaintenanceType = "inspection"
	MaintenanceTypeUpgrade     MaintenanceType = "upgrade"

	DeviceStatusOnline      DeviceStatus = "online"
	DeviceStatusOffline     DeviceStatus = "offline"
	DeviceStatusMaintenance DeviceStatus = "maintenance"



	InvoiceStatusUnpaid   InvoiceStatus = "unpaid"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusOverdue  InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"

	InvoiceTypeMonthly      InvoiceType = "monthly"
	InvoiceTypeVoucher      InvoiceType = "voucher"
	InvoiceTypeInstallation InvoiceType = "installation"
	InvoiceTypeOther        InvoiceType = "other"

	PaymentGatewayStatusPending PaymentGatewayStatus = "pending"
	PaymentGatewayStatusSuccess PaymentGatewayStatus = "success"
	PaymentGatewayStatusFailed  PaymentGatewayStatus = "failed"
	PaymentGatewayStatusExpired PaymentGatewayStatus = "expired"

	SegmentTypeBackbone     SegmentType = "Backbone"
	SegmentTypeDistribution SegmentType = "Distribution"
	SegmentTypeAccess       SegmentType = "Access"

	SegmentStatusActive      SegmentStatus = "active"
	SegmentStatusMaintenance SegmentStatus = "maintenance"
	SegmentStatusDamaged     SegmentStatus = "damaged"
	SegmentStatusInactive    SegmentStatus = "inactive"

	RequestStatusPending  RequestStatus = "pending"
	RequestStatusApproved RequestStatus = "approved"
	RequestStatusRejected RequestStatus = "rejected"

	NotificationTypeVoucherSold     NotificationType = "voucher_sold"
	NotificationTypePaymentReceived NotificationType = "payment_received"
	NotificationTypeBalanceUpdated  NotificationType = "balance_updated"
	NotificationTypeRequestApproved NotificationType = "request_approved"
	NotificationTypeRequestRejected NotificationType = "request_rejected"
)