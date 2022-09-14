package main

import (
	"fmt"
	"os"

	png "github.com/accrazed/png/src"
)

func main() {
	f, err := os.Open("png.png")
	if err != nil {
		panic(err)
	}
	t, err := png.NewTranscoder(f)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", t)
}
