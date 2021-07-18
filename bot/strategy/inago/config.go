package inago

import "fmt"

// Config is a parameter store of Inago Bot
type Config struct {
	volumeTriger float64
	scope        int64 // second
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

func (c *Config) Serialize() string {
	return fmt.Sprintf("%.0f-%d-%d-%t", c.volumeTriger, c.scope, c.settleTerm, c.reverse)

}
