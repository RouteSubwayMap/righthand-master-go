package dto

import (
	"encoding/json"
	"time"
)

// beijingLoc 北京时区(UTC+8)，前端传入的时间字符串均按此时区解析，
// 避免依赖操作系统 time.Local 在不同部署环境(如 Linux 默认 UTC)下导致 ±8 小时偏差。
var beijingLoc = time.FixedZone("CST", 8*60*60)

// CustomTime 自定义时间类型，支持多种格式解析
type CustomTime struct {
	time.Time
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return err
	}

	// 支持多种时间格式（按优先级排序）
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
		time.RFC3339,
	}

	// 始终按北京时区解析无时区信息的时间字符串；带时区的格式(如 RFC3339)会使用字符串自身的时区
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, beijingLoc); err == nil {
			ct.Time = t
			return nil
		}
	}

	return &time.ParseError{
		Layout:     "multiple formats tried",
		Value:      dateStr,
		Message:    "cannot parse time",
	}
}

// MarshalJSON 实现 json.Marshaler 接口
func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(ct.Time.Format(time.RFC3339))
}

// CustomTimePtr 自定义时间指针类型
type CustomTimePtr struct {
	*CustomTime
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (ctp *CustomTimePtr) UnmarshalJSON(data []byte) error {
	// 处理 null 值
	if string(data) == "null" {
		ctp.CustomTime = nil
		return nil
	}

	var ct CustomTime
	if err := ct.UnmarshalJSON(data); err != nil {
		return err
	}
	ctp.CustomTime = &ct
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (ctp *CustomTimePtr) MarshalJSON() ([]byte, error) {
	if ctp.CustomTime == nil {
		return json.Marshal(nil)
	}
	return ctp.CustomTime.MarshalJSON()
}