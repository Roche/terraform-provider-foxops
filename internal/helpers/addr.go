package helpers

func Addr[V any](v V) *V {
	return &v
}
