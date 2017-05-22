package slime

import (
	"fmt"
)

// logline formats a string with the given arguments, and then
// writes it to stdout with an extra newline.
func logline(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf(format+"\n", a...)
}
