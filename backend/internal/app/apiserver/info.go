package apiserver

import (
	"github.com/BronOS/secret-keeper/internal/pkg/db"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleInfo(logger *logrus.Logger, storage db.Interface, maxPinTries int8) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		if len(key) == 0 {
			logger.Error("empty key")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		secretSchema, err := storage.Get(key)
		if err != nil {
			logger.Error("secret not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger.Infof("secret info [%s] has been sent", key)

		writeJSON(w, map[string]interface{}{
			"expiration":   secretSchema.ExpTS,
			"pin_required": secretSchema.PinRequired,
			"tries_left":   maxPinTries - secretSchema.NumTries,
		}, http.StatusOK)
	}
}
