package benchmarker

type EncodableError struct {
	Message string
}

func (e EncodableError) Error() string {
	return e.Message
}

func encodeError(err error) *EncodableError {
	if err == nil {
		return nil
	}

	return &EncodableError{err.Error()}
}
