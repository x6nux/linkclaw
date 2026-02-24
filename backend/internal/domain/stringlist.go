package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// StringList 是字符串切片，在数据库中以 JSON 字符串形式存储于 TEXT 列。
// 实现 driver.Valuer 和 sql.Scanner，gorm 可直接用于 Raw/Scan。
type StringList []string

// Scan 从数据库读取（TEXT 列存储的 JSON 字符串）
func (s *StringList) Scan(value any) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("StringList.Scan: unsupported type")
	}
	str := strings.TrimSpace(string(b))
	if str == "" || str == "null" {
		*s = StringList{}
		return nil
	}
	var result []string
	if err := json.Unmarshal(b, &result); err != nil {
		*s = StringList{}
		return nil
	}
	*s = result
	return nil
}

// Value 序列化为 JSON 字符串写入数据库
func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Contains 判断是否包含指定字符串
func (s StringList) Contains(target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}
