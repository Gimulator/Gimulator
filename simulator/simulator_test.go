package simulator

import(
	_ "github.com/Gimulator/Gimulator/object"
	"github.com/Gimulator/Gimulator/storage"
	"reflect"
	"testing"
	"sync"
)

func TestNewSimulator(t *testing.T) {
	strg := storage.NewMemory()
	want := &Simulator{
		Mutex: sync.Mutex{},
		spreader: NewSpreader(),
		storage: strg,
	}

	got := NewSimulator(strg)

	if reflect.DeepEqual(got, want) {
		t.Logf(LogApproved(want, checkMark))
	} else {
		t.Errorf(LogFailed(got, want, ballotX))
	}
}