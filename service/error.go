package service

type Error interface {
	error
	Unwrap() error
	Status() int
}

type statusError struct {
	err        error
	statusCode int
}

func NewError(err error, statusCode int) error {
	if err == nil {
		return nil
	}
	return statusError{
		err:        err,
		statusCode: statusCode,
	}
}

func (e statusError) Error() string {
	return e.err.Error()
}

func (e statusError) Unwrap() error {
	return e.err
}

func (e statusError) Status() int {
	return e.statusCode
}
