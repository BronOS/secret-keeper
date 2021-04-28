package apiserver

import (
	"github.com/BronOS/secret-keeper/internal/pkg/db"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleDelete(logger *logrus.Logger, storage db.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		if len(key) == 0 {
			logger.Error("empty key")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err := storage.Get(key)
		if err != nil {
			logger.Error("secret not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if derr := storage.Delete(key); derr != nil {
			logger.Errorf("failed to delete secret: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Infof("secret [%s] has been deleted successfully", key)
		w.WriteHeader(http.StatusOK)
	}
}
