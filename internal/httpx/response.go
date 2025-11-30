package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/retail-core/sales-service/internal/errors"
	"github.com/retail-core/sales-service/internal/logger"
	"go.uber.org/zap"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, appError error) {
	err, ok := errors.IsAppError(appError)
	if !ok {
		logger.L().Error("Internal Server Error", zap.Error(appError))
		err = errors.Internal("Ooops!!! Something wrong happened")
		
	}
	WriteJSON(w, err.Status, err)
}
