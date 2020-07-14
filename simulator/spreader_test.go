package simulator

import (
	_ "fmt"
	"testing"
	_ "github.com/Gimulator/Gimulator/object"
	"github.com/sirupsen/logrus"
	"reflect"
)

func TestNewSpreader(t *testing.T) {
	want := &spreader {
		watchers : make(map[string]watcher),
		log : logrus.WithField("entity", "spreader"),
	}

	t.Logf("Given the need to test NewSpreader function of watcher type.")

	got := NewSpreader()

	if reflect.DeepEqual(got, want) {
		t.Logf(LogApproved(want, checkMark))
	} else {
		t.Errorf(LogFailed(got, want, ballotX))
	}
}

