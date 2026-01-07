package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-V") {
		fmt.Println("woeusb", version)
		return
	}
	fmt.Println("woeusb", version)
	fmt.Println("Usage: woeusb [--device | --partition] [options] <source> <target>")
}
