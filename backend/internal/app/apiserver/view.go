package apiserver

import (
	"github.com/BronOS/secret-keeper/internal/pkg/db"
	"github.com/BronOS/secret-keeper/internal/pkg/security"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleView(logger *logrus.Logger, storage db.Interface, maxPinTries int8, cipher security.CipherInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		if len(key) == 0 {
			logger.Error("empty key")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		pin := r.URL.Query().Get("pin")

		secretSchema, err := storage.Get(key)
		if err != nil {
			logger.Error("secret not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		secret, err := cipher.Decrypt(pin, secretSchema.Pin, secretSchema.Secret)
		if err != nil {
			logger.Error(err)

			if (secretSchema.NumTries + 1) >= maxPinTries {
				logger.Errorf("secret [%s] riched max tries", key)
				if derr := storage.Delete(key); derr != nil {
					logger.Errorf("failed to delete secret: %v", err)
				}

				w.WriteHeader(http.StatusNotFound)
				return
			}

			if ierr := storage.IncNumTries(key); ierr != nil {
				logger.Errorf("failed to increment num tries: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			writeJSON(w, map[string]interface{}{
				"tries_left": maxPinTries - secretSchema.NumTries - 1,
			}, http.StatusForbidden)
			return
		}

		logger.Infof("secret [%s] has been viewed and deleted successfully", key)

		if derr := storage.Delete(key); derr != nil {
			logger.Errorf("failed to delete secret: %v", err)
		}

		writeJSON(w, map[string]interface{}{
			"secret": secret,
		}, http.StatusOK)
	}
}
