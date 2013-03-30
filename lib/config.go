package lib

// WFSD config object
type Config struct {
	Paths []Path
}

// Path entry 
type Path struct {
	Prefix      string
	Directory   string
	StripPrefix bool
}

// Return true if the root path "/" is present in this config
func (c *Config) IsRootSet() bool {
	if c == nil {
		return false
	}
	for _, path := range c.Paths {
		if path.Prefix == "/" {
			return true
		}
	}
	return false
}
