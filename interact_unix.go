// +build !windows

package main

import (
	"bufio"
	"os"
	"fmt"
)

var input *bufio.Reader

func init() {
	input = bufio.NewReader(os.Stdin)
}

func ask(question string) bool {
	fmt.Print(question + "\n[Y/n] ")
	line, _ := input.ReadString('\n')
	char := line[0]
	return char == 'y' || char == 'Y' || char == '\n'
}

func info(message string) {
	fmt.Print(message)
	input.ReadString('\n')
}
