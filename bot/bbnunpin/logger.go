package bbnunpin

import (
	"fmt"
	"time"
)

func Log(msg ...interface{}) {
	t := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(t, msg)
}
