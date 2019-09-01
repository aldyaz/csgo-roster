package response

import (
	"encoding/json"
	"net/http"
)

func JSON(res http.ResponseWriter, statusCode int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	json.NewEncoder(res).Encode(data)
}
