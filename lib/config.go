package lib

type Config struct {
	Paths []Path
}

type Path struct {
	Prefix      string
	Directory   string
	StripPrefix bool
}
