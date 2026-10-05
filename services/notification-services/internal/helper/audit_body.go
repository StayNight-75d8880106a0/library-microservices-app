package helper

import "encoding/json"

const MaxAuditBodySize = 32 * 1024 // 32 KB

func SafeJSONBody(raw []byte) json.RawMessage {

	if len(raw) == 0 {
		return nil
	}

	if len(raw) > MaxAuditBodySize {
		return json.RawMessage(`{"_truncated":true}`)
	}

	if !json.Valid(raw) {
		return json.RawMessage(`{"_invalid":true}`)
	}

	return json.RawMessage(raw)

}
