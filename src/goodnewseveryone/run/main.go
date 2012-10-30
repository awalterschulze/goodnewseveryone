package main 

import (
	"goodnewseveryone"
	"goodnewseveryone/web"
	"flag"
)

//windows 7 registry fix
//http://alan.lamielle.net/2009/09/03/windows-7-nonpaged-pool-srv-error-2017

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