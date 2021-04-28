package passwords

type GeneratorInterface interface {
	Generate() (string, error)
}
