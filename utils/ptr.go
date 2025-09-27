package utils

import "time"

func Int32Ptr(v int32) *int32 {
	return &v
}

func Int64Ptr(v int64) *int64 {
	return &v
}

func StringPtr(v string) *string {
	return &v
}

func TimePtr(v time.Time) *time.Time {
	return &v
}