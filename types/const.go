package types

type Role string

const (
	DirectorRole Role = "director"
	OperatorRole Role = "operator"
	MasterRole   Role = "master"
)

type Method string

const (
	GetMethod       Method = "get"
	GetAllMethod    Method = "get-all"
	PutMethod       Method = "put"
	DeleteMethod    Method = "delete"
	DeleteAllMethod Method = "delete-all"
	WatchMethod     Method = "watch"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusFailed  Status = "failed"
	StatusUnknown Status = "unknown"
)
