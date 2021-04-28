package db

type Interface interface {
	Connect() error
	Disconnect() error
	Set(s *SecretSchema) error
	Get(key string) (*SecretSchema, error)
	IncNumTries(key string) error
	Delete(key string) error
}
