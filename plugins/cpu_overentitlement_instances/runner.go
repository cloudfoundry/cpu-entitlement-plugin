package cpu_overentitlement_instances

import "code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"

//go:generate counterfeiter . Reporter

type Reporter interface {
	OverEntitlementInstances() (reporter.Report, error)
}

//go:generate counterfeiter . Renderer

type Renderer interface {
	Render(reporter.Report) error
}

type Runner struct {
	reporter Reporter
	renderer Renderer
}

func NewRunner(reporter Reporter, renderer Renderer) *Runner {
	return &Runner{
		reporter: reporter,
		renderer: renderer,
	}
}

func (r *Runner) Run() error {
	report, err := r.reporter.OverEntitlementInstances()
	if err != nil {
		return err
	}

	return r.renderer.Render(report)
}
