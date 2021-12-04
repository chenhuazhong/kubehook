package utils

type TypeError struct {
	Message string
}

func (err *TypeError) Error() string {
	return err.Message
}

func (err *TypeError) String() string {
	return err.Message
}
