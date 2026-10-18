package strutil

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var (
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

// ExtractProxyURIName 从代理 URI（支持 vmess:// base64 JSON 的 ps 字段及 #fragment）提取节点名称。
func ExtractProxyURIName(rawURI string) string {
	rawURI = strings.TrimSpace(rawURI)
	if strings.HasPrefix(strings.ToLower(rawURI), "vmess://") {
		b64Str := rawURI[len("vmess://"):]
		if idx := strings.IndexAny(b64Str, "?#"); idx >= 0 {
			b64Str = b64Str[:idx]
		}
		if b, err := base64.StdEncoding.DecodeString(PadBase64(b64Str)); err == nil {
			var payload map[string]any
			if json.Unmarshal(b, &payload) == nil {
				if ps, ok := payload["ps"].(string); ok && strings.TrimSpace(ps) != "" {
					return strings.TrimSpace(ps)
				}
			}
		}
	}
	if idx := strings.LastIndex(rawURI, "#"); idx >= 0 {
		escapedName := rawURI[idx+1:]
		if dec, err := url.QueryUnescape(escapedName); err == nil {
			return strings.TrimSpace(dec)
		}
		return strings.TrimSpace(escapedName)
	}
	return ""
}

