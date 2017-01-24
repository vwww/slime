package slime

import (
	"encoding/json"
	"net/http"
)

func writeJSON(res http.ResponseWriter, obj interface{}, okPrefix, errMsg []byte) {
	// Try to encode as JSON
	b, err := json.Marshal(obj)
	if err != nil {
		res.Write(errMsg)
		return
	}

	// Set headers
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "no-cache, no-store")

	// Write output
	res.Write(okPrefix)
	res.Write(b)
}
