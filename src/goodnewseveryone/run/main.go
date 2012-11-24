//Copyright 2012 Walter Schulze
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

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