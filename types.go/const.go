package types

type Role string

const (
	DirectorRole Role = "director"
	OperatorRole Role = "operator"
	MasterRole   Role = "master"
)

type Method string

const (
	GetMethod       Method = "get-method"
	GetAllMethod    Method = "get-all-method"
	PutMethod       Method = "put-method"
	DeleteMethod    Method = "delete-method"
	DeleteAllMethod Method = "delete-all-method"
	WatchMethod     Method = "watch-method"
)
