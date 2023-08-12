package main

import (
	"fmt"
	"os/exec"
)

func executeLs() ([]byte, error) {
	cmd := exec.Command("ls", "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	out, err := executeLs()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Output: %s\n", out)
}
