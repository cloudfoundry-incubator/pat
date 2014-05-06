package benchmarker

import "encoding/json"

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

func (e *EncodableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

func (e *EncodableError) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, e.Message)
}
