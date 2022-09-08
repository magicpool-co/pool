package types

import (
	"time"
)

func TimePtr(v time.Time) *time.Time {
	return &v
}

func TimeValue(v *time.Time) time.Time {
	if v != nil {
		return *v
	}

	return time.Time{}
}

func StringPtr(v string) *string {
	return &v
}

func StringValue(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

func IntPtr(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

func Int32Ptr(v int32) *int32 {
	return &v
}

func Int32alue(v *int32) int32 {
	if v != nil {
		return *v
	}

	return 0
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}

func UintPtr(v uint) *uint {
	return &v
}

func UintValue(v *uint) uint {
	if v != nil {
		return *v
	}

	return 0
}

func Uint32Ptr(v uint32) *uint32 {
	return &v
}

func Uint32Value(v *uint32) uint32 {
	if v != nil {
		return *v
	}

	return 0
}

func Uint64Ptr(v uint64) *uint64 {
	return &v
}

func Uint64Value(v *uint64) uint64 {
	if v != nil {
		return *v
	}

	return 0
}

func Float32Ptr(v float32) *float32 {
	return &v
}

func Float32Value(v *float32) float32 {
	if v != nil {
		return *v
	}

	return 0
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func Float64Value(v *float64) float64 {
	if v != nil {
		return *v
	}

	return 0
}

func BoolPtr(v bool) *bool {
	return &v
}

func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}

	return false
}
