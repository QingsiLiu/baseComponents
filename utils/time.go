package utils

import (
	"context"
	"fmt"
	"time"
)

const (
	// 常用时间格式
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
	RFC3339Format  = time.RFC3339
	UnixFormat     = "1136239445"
)

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// NowUnix 获取当前Unix时间戳
func NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli 获取当前毫秒级Unix时间戳
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatNow 格式化当前时间
func FormatNow(layout string) string {
	return time.Now().Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseTimeInLocation 在指定时区解析时间字符串
func ParseTimeInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// ToUTC 转换为UTC时间
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ToLocal 转换为本地时间
func ToLocal(t time.Time) time.Time {
	return t.Local()
}

// ToTimezone 转换到指定时区
func ToTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, err
	}
	return t.In(loc), nil
}

// StartOfDay 获取一天的开始时间 (00:00:00)
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束时间 (23:59:59.999999999)
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek 获取一周的开始时间 (周一 00:00:00)
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}
	days := weekday - 1
	return StartOfDay(t.AddDate(0, 0, -days))
}

// EndOfWeek 获取一周的结束时间 (周日 23:59:59.999999999)
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth 获取一个月的开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取一个月的结束时间
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// StartOfYear 获取一年的开始时间
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 获取一年的结束时间
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// DiffInDays 计算两个时间相差的天数
func DiffInDays(t1, t2 time.Time) int {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	return int(t2.Sub(t1).Hours() / 24)
}

// DiffInHours 计算两个时间相差的小时数
func DiffInHours(t1, t2 time.Time) int {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	return int(t2.Sub(t1).Hours())
}

// DiffInMinutes 计算两个时间相差的分钟数
func DiffInMinutes(t1, t2 time.Time) int {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}
	return int(t2.Sub(t1).Minutes())
}

// IsToday 判断是否是今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsYesterday 判断是否是昨天
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// IsTomorrow 判断是否是明天
func IsTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day()
}

// IsWeekend 判断是否是周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWorkday 判断是否是工作日
func IsWorkday(t time.Time) bool {
	return !IsWeekend(t)
}

// Age 计算年龄
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// 如果还没到生日，年龄减1
	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}

// HumanDuration 将时间间隔转换为人类可读的格式
func HumanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f分钟", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
	return fmt.Sprintf("%.1f天", d.Hours()/24)
}

// Sleep 休眠指定时间
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// Timeout 创建一个超时上下文
func Timeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
