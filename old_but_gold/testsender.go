package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	pipe, err := os.Open("/tmp/my_pipe")
	if err != nil {
		fmt.Println("Error opening pipe:", err)
		return
	}
	defer pipe.Close()

	reader := bufio.NewReader(pipe)
	message, _ := reader.ReadString('\n')
	fmt.Println("Received message:", message)
}

