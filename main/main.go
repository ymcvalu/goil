package main

import (
	"fmt"
	"path"
)

func main() {
	paths := []string{
		"www/",
		"/hao",
		"/123/",
	}
	url := path.Join(paths...)
	fmt.Println(url)
}
