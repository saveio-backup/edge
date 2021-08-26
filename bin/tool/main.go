package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"time"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	var path = flag.String("path", ".", "output file path")
	var size = flag.Int("size", 1, "size in KB")
	flag.Parse()
	data := make([]byte, *size)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(*path, data, 0666); err != nil {
		panic(err)
	}
}
