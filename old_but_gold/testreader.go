package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	pipe, err := os.OpenFile("/tmp/my_pipe", os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		fmt.Println("Error opening pipe:", err)
		return
	}
	defer pipe.Close()

	fmt.Print("Enter the string to send: ")
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	pipe.WriteString(input)
	fmt.Println("Message sent!")
}
