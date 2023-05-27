package contracts

import (
	"testing"
)

func TestABISigs(t *testing.T) {
	for name, event := range SwapCreatorParsedABI.Events {
		t.Logf("%s: %s", name, event.Sig)
	}
	t.Log()
	for name, method := range SwapCreatorParsedABI.Methods {
		t.Logf("%s: %s", name, method.Sig)
	}
}
