package inago

// Config is a parameter store of Inago Bot
type Config struct {
	scope        int64 // second
	volumeTriger float64
	settleTerm   int64
	reverse      bool // buyTriger検知でshort/sellTriger検知でlong
}

func NewConfig(scope int64, volumeTriger float64, settleTerm int64, reverse bool) Config {
	return Config{
		scope:        scope,
		volumeTriger: volumeTriger,
		settleTerm:   settleTerm,
		reverse:      reverse,
	}
}
