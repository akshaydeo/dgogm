package dgorm

import "fmt"

func Debug(f string, p ...interface{}) {
	fmt.Printf(f, p)
}

func Error(f string, p ...interface{}) {
	fmt.Printf(f, p)
}
