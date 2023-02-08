package dean

type Dispatch struct {
	Path string
	Id   string
}

type Announce struct {
	Dispatch
	Model string
	Name  string
}
