package main

import "os"

func main() {
	os.Exit(0) // want "forbidden direct call os.Exit"
}
