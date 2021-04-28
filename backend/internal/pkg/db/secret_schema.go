package db

import "time"

type SecretSchema struct {
	Key         string `bson:"key"`
	Secret      []byte `bson:"secret"`
	Pin         []byte `bson:"pin"`
	ExpTS       int64  `bson:"exp_ts"`
	Ttl         int64  `bson:"ttl"`
	NumTries    int8   `bson:"num_tries"`
	PinRequired bool   `bson:"pin_required"`
}

func NewSecretSchema(key string, secret, pin []byte, pinRequired bool, ttl int64) *SecretSchema {
	return &SecretSchema{
		Key:         key,
		Secret:      secret,
		Pin:         pin,
		ExpTS:       time.Now().Unix() + ttl,
		Ttl:         ttl,
		NumTries:    0,
		PinRequired: pinRequired,
	}
}
