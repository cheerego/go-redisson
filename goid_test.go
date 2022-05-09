package godisson

import (
	"testing"
)

func Test_gid(t *testing.T) {
	goid, err := gid()
	if err != nil {
		t.Errorf("gid() error = %v", err)
		return
	}
	t.Log(goid)
}
