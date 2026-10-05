package domain

type OperationType string

const (
	OpTypeStripeDeposit OperationType = "STRIPE_DEPOSIT"
	OpTypeServiceUsage  OperationType = "SERVICE_USAGE"
	OpTypeRefund        OperationType = "REFUND"
)

type DepositStatus string

const (
	StatusPending DepositStatus = "pending"
	StatusSuccess DepositStatus = "success"
	StatusFailed  DepositStatus = "failed"
)
