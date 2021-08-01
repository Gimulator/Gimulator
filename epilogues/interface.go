package epilogues

import "github.com/Gimulator/protobuf/go/api"

type Epilogue interface {
	Write(result *api.Result) error
	Test() error
}
