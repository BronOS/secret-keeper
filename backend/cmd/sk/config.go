package main

type Config struct {
	Server struct {
		Bind string `yaml:"bind"`
	} `yaml:"server"`

	Logger struct {
		Level string `yaml:"level"`
	} `yaml:"logger"`

	PasswordGenerator struct {
		Length     int `yaml:"length"`
		NumDigits  int `yaml:"num_digits"`
		NumSymbols int `yaml:"num_symbols"`
	} `yaml:"password_generator"`

	Database struct {
		Addr string `yaml:"addr"`
		Port int32  `yaml:"port"`
		Name string `yaml:"name"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"database"`

	Security struct {
		CipherKey    string `yaml:"cipher_key"`
		MaxPinTries  int8   `yaml:"max_pin_tries"`
		MaxBodyBytes int64  `yaml:"max_body_bytes"`
		MaxTTL       int64  `yaml:"max_ttl"`
	} `yaml:"security"`
}
