package fsm

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Lint(t *testing.T, fsm FSM) {
	l := linter{fsm: fsm}
	l.checkStates(t)
	l.checkNames(t)
}

type linter struct {
	fsm FSM
}

func (l *linter) checkStates(t *testing.T) {
	sfType := reflect.TypeOf(StateFunction(nil))
	fsmTyp := reflect.TypeOf(l.fsm)
	fsmVal := reflect.ValueOf(l.fsm)
	sfs := l.fsm.StateFunctions()
	for i := 0; i < fsmTyp.NumMethod(); i++ {
		mTyp := fsmTyp.Method(i)
		mVal := fsmVal.Method(i)
		if mVal.CanConvert(sfType) {
			if _, ok := sfs[mTyp.Name]; !ok {
				t.Logf("missing state function %q", mTyp.Name)
				t.Fail()
			}
		}
	}
}

func (l *linter) checkNames(t *testing.T) {
	for got, sf := range l.fsm.StateFunctions() {
		if want := l.getName(sf); got != want {
			t.Logf("unexpected state function name, want %q, got %q", want, got)
			t.Fail()
		}
	}
}

func (*linter) getName(f StateFunction) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	i := strings.LastIndexByte(name, '.')
	if i >= 0 {
		name = name[i+1:]
	}
	name = strings.TrimSuffix(name, "-fm")
	return name
}
