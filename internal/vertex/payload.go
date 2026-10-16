package vertex

import (
	"crypto/rand"
	"fmt"
	"math/big"

)

// batchGraphql 请求的固定包壳常量（逐字节对齐 PoC body.json / vertex_client）。
const (
	querySignature = "2/l8eCsMMY49imcDQ/lwwXyL8cYtTjxZBF2dNqy69LodY="
	operationName  = "StreamGenerateContentAnonymous"
)

func randomPageViewID() int64 {
	minVal := int64(1000000000000000)
	maxVal := int64(9000000000000000)
	n, err := rand.Int(rand.Reader, big.NewInt(maxVal))
	if err != nil {
		return minVal
	}
	return minVal + n.Int64()
}

func randomTrackingID() string {
	digits := make([]byte, 16)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err == nil {
			digits[i] = '0' + byte(n.Int64())
		} else {
			digits[i] = '0'
		}
	}
	return "d" + string(digits)
}

func randomUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant to RFC4122
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%X-%X-%X-%X-%X",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// buildRequestPayload 构建发往上游的完整请求体：
// 纯原生零干涉直通，客户端发送的原生 Gemini payload 原封不动放入 variables，
// 仅注入 model, region=global, recaptchaToken，绝对不截断参数或伪造回合。
func buildRequestPayload(model string, geminiPayload map[string]any, recaptchaToken string) map[string]any {
	vars := make(map[string]any, len(geminiPayload)+3)
	for k, v := range geminiPayload {
		vars[k] = v
	}
	vars["model"] = model
	vars["region"] = "global"
	vars["recaptchaToken"] = recaptchaToken

	trackingID := randomTrackingID()
	return map[string]any{
		"requestContext": map[string]any{
			"clientVersion":    "boq_cloud-boq-clientweb-vertexaistudio_20260630.00_p0",
			"pagePath":         "/agent-platform/studio/multimodal",
			"pageViewId":       randomPageViewID(),
			"trackingId":       trackingID,
			"backendOverrides": map[string]any{},
			"clientSessionId":  randomUUID(),
			"selectedPurview":  map[string]any{},
			"jurisdiction":     "global",
			"localizationData": map[string]any{
				"locale":   "zh_CN",
                "timezone": "Asia/Hong_Kong",
			},
		},
		"querySignature": querySignature,
		"operationName":  operationName,
		"variables":      vars,
	}
}
