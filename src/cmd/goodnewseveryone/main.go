package main 

import (
	"goodnewseveryone"
	"flag"
)

var (
	configLocation = "."
)

func main() {
	var configLocation = flag.String("config", ".", "folder where all the config files are located")
	flag.Parse()
	goodnewseveryone.Main(*configLocation)
}