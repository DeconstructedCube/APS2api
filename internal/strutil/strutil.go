package strutil

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

var (
	camelRe    = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	reqCounter uint64
)

// ReqID 生成权威 24 字符十六进制请求 ID。
func ReqID() string {
	var buf [12]byte
	if _, err := cryptorand.Read(buf[:]); err != nil {
		now := time.Now().UnixNano()
		count := atomic.AddUint64(&reqCounter, 1)
		var fallback [12]byte
		fallback[0] = byte(now >> 56)
		fallback[1] = byte(now >> 48)
		fallback[2] = byte(now >> 40)
		fallback[3] = byte(now >> 32)
		fallback[4] = byte(now >> 24)
		fallback[5] = byte(now >> 16)
		fallback[6] = byte(now >> 8)
		fallback[7] = byte(now)
		fallback[8] = byte(count >> 24)
		fallback[9] = byte(count >> 16)
		fallback[10] = byte(count >> 8)
		fallback[11] = byte(count)
		return hex.EncodeToString(fallback[:])
	}
	return hex.EncodeToString(buf[:])
}

// PadBase64 替换 URL-safe base64 字符并补齐 '=' 填充位。
func PadBase64(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "-", "+"), "_", "/")
	if pad := len(s) % 4; pad != 0 {
		s += strings.Repeat("=", 4-pad)
	}
	return s
}

// FirstNonEmpty 返回参数列表中第一个非空字符串，全部为空则返回空字符串。
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// SnakeToCamel 将 snake_case 转为 camelCase。
func SnakeToCamel(s string) string {
	if !strings.Contains(s, "_") {
		return s
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		b.WriteString(pyTitle(p))
	}
	return b.String()
}

// CamelToSnake 将 camelCase 转为 snake_case。
func CamelToSnake(s string) string {
	return strings.ToLower(camelRe.ReplaceAllString(s, "${1}_${2}"))
}

func pyTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
