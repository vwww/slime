package slime

import (
	"fmt"
)

func logline(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format+"\n", a...)
}
