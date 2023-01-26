package env

import "errors"

// ErrInvalidString env not is set
var ErrEnvNotIsset = errors.New("not is set")

// ErrEnvConvert env conversion error
var ErrEnvConvert = errors.New("conversion error")

var _ error = (*envError)(nil)

type envError struct {
	Err       error
	ErrParent error
	EnvName   string
	EnvValue  string
}

func NewError(err error, envName, envValue string) *envError {
	return &envError{Err: err, EnvName: envName, EnvValue: envValue}
}

func (e *envError) Unwrap() error { return e.Err }

func (e *envError) Error() string {
	if e == nil {
		return "<nil>"
	}
	s := "env name" + e.EnvName
	s += "=" + e.EnvValue
	s += " " + e.Err.Error()
	if e.ErrParent != nil {
		s += " error: " + e.ErrParent.Error()
	}
	return s
}

func (e *envError) AddParentErr(err error) *envError {
	e.ErrParent = err
	return e
}
