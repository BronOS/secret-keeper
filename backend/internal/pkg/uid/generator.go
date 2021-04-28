package uid

import (
	"github.com/segmentio/ksuid"
)

type Generator struct {
}

func (g *Generator) Generate() string {
	return ksuid.New().String()
}

func NewGenerator() *Generator {
	return &Generator{}
}
