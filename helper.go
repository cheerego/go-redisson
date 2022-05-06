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

func IntPtr(v int) *int {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func UintPtr(v uint) *uint {
	return &v
}

func Uint64Ptr(v uint64) *uint64 {
	return &v
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func BoolPtr(v bool) *bool {
	return &v
}

func StringPtr(v string) *string {
	return &v
}

func StringValues(ptrs []*string) []string {
	values := make([]string, len(ptrs))
	for i := 0; i < len(ptrs); i++ {
		if ptrs[i] != nil {
			values[i] = *ptrs[i]
		}
	}
	return values
}

func IntPtrs(vals []int) []*int {
	ptrs := make([]*int, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func Int64Ptrs(vals []int64) []*int64 {
	ptrs := make([]*int64, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func UintPtrs(vals []uint) []*uint {
	ptrs := make([]*uint, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func Uint64Ptrs(vals []uint64) []*uint64 {
	ptrs := make([]*uint64, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func Float64Ptrs(vals []float64) []*float64 {
	ptrs := make([]*float64, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func BoolPtrs(vals []bool) []*bool {
	ptrs := make([]*bool, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func StringPtrs(vals []string) []*string {
	ptrs := make([]*string, len(vals))
	for i := 0; i < len(vals); i++ {
		ptrs[i] = &vals[i]
	}
	return ptrs
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}
