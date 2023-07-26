package dean

type Runner struct {
	thinger  Thinger
	bus      *Bus
	injector *Injector
}

func NewRunner(thinger Thinger) *Runner {
	var r Runner

	r.thinger = thinger

	r.bus = NewBus("runner bus", nil, nil)
	r.injector = NewInjector("runner injector", r.bus)

	return &r
}

func (r *Runner) Run() {
	r.thinger.SetFlag(ThingFlagMetal)
	r.thinger.Run(r.injector)
}
