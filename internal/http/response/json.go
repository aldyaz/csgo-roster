package response

import (
	"encoding/json"
	"net/http"
)

func JSON(res http.ResponseWriter, statusCode int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	err := json.NewEncoder(res).Encode(data)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}
}
