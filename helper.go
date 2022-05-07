package godisson

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GetGoid() (int64, error) {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)
	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return 0, errors.Wrapf(err, "can not get goroutine id: %v", idField)
	}
	return int64(id), nil
}

func GetLockName(uuid string) (string, error) {
	goid, err := GetGoid()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", uuid, goid), nil
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}
