package utils

type APIResource interface {
	String()
	Update(func(instance *APIResource, validata ...interface{}))
}
