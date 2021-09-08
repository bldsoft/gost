package utils

type ContextKey struct {
	Name string
}

func (k *ContextKey) String() string {
	return "Gost ctx " + k.Name
}
