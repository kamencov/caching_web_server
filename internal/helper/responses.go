package helper

import (
	"caching_web_server/internal/models"
	"encoding/json"
	"net/http"
)

// WriteResponse - запись ответа
func WriteResponse(w http.ResponseWriter, code int, resp models.Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}

}

// OkResponse - успешный ответ
func OkResponse(w http.ResponseWriter, resp interface{}) {
	WriteResponse(w, http.StatusOK, models.Envelope{
		Response: resp})
}

// FailResponse - ответ с ошибкой
func FailResponse(w http.ResponseWriter, code int, message string) {
	WriteResponse(w, code, models.Envelope{
		Error: &models.ErrPayload{
			Code: code,
			Text: message,
		},
	})
}

// OkDataResponse - ответ с данными
func OkDataResponse(w http.ResponseWriter, data interface{}) {
	WriteResponse(w, http.StatusOK, models.Envelope{
		Data: data,
	})
}
