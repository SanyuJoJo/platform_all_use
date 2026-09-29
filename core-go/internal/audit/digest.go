package audit
import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
)
// ParamsDigest 计算 params 的 SHA-256 摘要。
//
// 说明：为避免同一 params 因 map 顺序不同产生不同摘要，
// 先 Marshal（encoding/json 对 map 稳定按 key 排序）再 hash。
func ParamsDigest(params map[string]interface{}) string {
    if params == nil {
        return "sha256:unknown"
    }
    data, err := json.Marshal(params)
    if err != nil {
        return "sha256:unknown"
    }
    h := sha256.Sum256(data)
    return "sha256:" + hex.EncodeToString(h[:])
}
