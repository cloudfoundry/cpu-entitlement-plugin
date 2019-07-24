package result

import "code.cloudfoundry.org/cli/cf/terminal"

type Result struct {
	isFailure      bool
	errorMessage   string
	warningMessage string
}

func (r Result) IsFailure() bool {
	return r.isFailure
}

func Failure(errorMessage string) Result {
	return Result{
		isFailure:    true,
		errorMessage: errorMessage,
	}
}

func FailureFromError(err error) Result {
	return Result{
		isFailure:    true,
		errorMessage: err.Error(),
	}
}

func Success() Result {
	return Result{}
}

func (r Result) WithWarning(warning string) Result {
	return Result{
		isFailure:      r.isFailure,
		errorMessage:   r.errorMessage,
		warningMessage: warning,
	}
}

func (r Result) WriteTo(ui terminal.UI) {
	if r.errorMessage != "" {
		ui.Failed(r.errorMessage)
	}

	if r.warningMessage != "" {
		ui.Warn(r.warningMessage)
	}
}
