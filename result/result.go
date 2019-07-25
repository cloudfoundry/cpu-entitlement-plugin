package result

type Result struct {
	IsFailure      bool
	ErrorMessage   string
	WarningMessage string
}

func Failure(errorMessage string) Result {
	return Result{
		IsFailure:    true,
		ErrorMessage: errorMessage,
	}
}

func FailureFromError(err error) Result {
	return Result{
		IsFailure:    true,
		ErrorMessage: err.Error(),
	}
}

func Success() Result {
	return Result{}
}

func (r Result) WithWarning(warning string) Result {
	return Result{
		IsFailure:      r.IsFailure,
		ErrorMessage:   r.ErrorMessage,
		WarningMessage: warning,
	}
}
