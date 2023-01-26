package env

type EnvInterface interface {
	Get(name string) string
	GetMust(name string) string
	GetMustInt(name string) int
}
