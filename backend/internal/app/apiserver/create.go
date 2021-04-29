package apiserver

import (
	"encoding/json"
	"github.com/BronOS/secret-keeper/internal/pkg/db"
	"github.com/BronOS/secret-keeper/internal/pkg/passwords"
	"github.com/BronOS/secret-keeper/internal/pkg/security"
	"github.com/BronOS/secret-keeper/internal/pkg/uid"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleCreate(logger *logrus.Logger, storage db.Interface, pg passwords.GeneratorInterface, kg uid.GeneratorInterface, cipher security.CipherInterface, maxBodySize, maxTTL int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		isSecretGeneratedFlag := false
		dto := &struct {
			Secret string `json:"secret"`
			Ttl    int64  `json:"ttl"`
			Pin    string `json:"pin"`
		}{}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		if err = json.NewDecoder(r.Body).Decode(dto); err != nil {
			logger.Errorf("failed to decode JSON request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if dto.Ttl <= 0 || dto.Ttl > maxTTL {
			dto.Ttl = maxTTL
		}

		if len(dto.Secret) == 0 {
			dto.Secret, err = pg.Generate()
			if err != nil {
				logger.Errorf("failed to generate secret: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			isSecretGeneratedFlag = true
		}

		key := kg.Generate()
		pinHash, secretHash, err := cipher.Encrypt(dto.Pin, dto.Secret)
		isPinRequiredFlag := len(dto.Pin) > 0

		if err != nil {
			logger.Errorf("failed to encrypt secret: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		secretSchema := db.NewSecretSchema(key, secretHash, pinHash, isPinRequiredFlag, dto.Ttl)

		if err = storage.Set(secretSchema); err != nil {
			logger.Errorf("failed to store secret: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Infof("secret [%s] has been created successfully", key)

		response := map[string]interface{}{
			"key":          key,
			"expiration":   secretSchema.ExpTS,
			"pin_required": isPinRequiredFlag,
		}

		if isSecretGeneratedFlag {
			response["secret"] = dto.Secret
		}

		writeJSON(w, response, http.StatusCreated)
	}
}
