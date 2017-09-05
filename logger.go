package dgorm

import "fmt"

func Debug(f string, p ...interface{}) {
	fmt.Printf(fmt.Sprintf(f, p...))
}

func Error(f string, p ...interface{}) {
	fmt.Printf(fmt.Sprintf(f, p...))
}
