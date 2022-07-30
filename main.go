package main

import (
	"fmt"

	png "github.com/accrazed/png/src"
)

func main() {
	t, err := png.NewTranscoder("png.png")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", t)
}
