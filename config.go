package main

type Config struct {
	Port string
	Paths []Path
}

type Path struct {
	Prefix string
	Directory string
	StripPerfix bool
}
