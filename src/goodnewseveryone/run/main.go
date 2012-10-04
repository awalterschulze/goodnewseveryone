package main 

import (
	"goodnewseveryone"
	"goodnewseveryone/web"
	"flag"
)

var (
	configLocation = "."
)

func main() {
	var configLocation = flag.String("config", ".", "folder where all the config files are located")
	flag.Parse()
	gne := goodnewseveryone.ConfigToGNE(*configLocation)
	go gne.Start()
	web.Serve(gne)
}