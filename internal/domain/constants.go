package domain

type OperationType string

const (
	OpTypeStripeDeposit OperationType = "STRIPE_DEPOSIT"
	OpTypeServiceUsage  OperationType = "SERVICE_USAGE"
	OpTypeRefund        OperationType = "REFUND"
)
