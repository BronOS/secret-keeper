package passwords

import "github.com/sethvargo/go-password/password"

type Generator struct {
	c *Config
}

func (g *Generator) Generate() (string, error) {
	return password.Generate(g.c.Length, g.c.NumDigits, g.c.NumSymbols, false, false)
}

func NewGenerator(c *Config) *Generator {
	return &Generator{c: c}
}
