package types

type Role string

const (
	DirectorRole Role = "director"
	OperatorRole Role = "operator"
	MasterRole   Role = "master"
)

type Method string

const (
	Get       Method = "get-method"
	GetAll    Method = "get-all-method"
	Put       Method = "put-method"
	Delete    Method = "delete-method"
	DeleteAll Method = "delete-all-method"
	Watch     Method = "watch-method"
)
