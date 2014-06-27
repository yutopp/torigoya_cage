//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package main

import (
	"log"
	"os"
	"io/ioutil"

	"yutopp/torigoya/cage"

	"gopkg.in/v1/yaml"
)


type Config map[string]struct {
	Host		string
	Port		int
	Host_user	string
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicf("Error (%v)\n", err)
	}

	log.Printf("%s\n", cwd)

	config_bytes, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Panic("There is no \"config.yml\" file...")
	}

	//
	config := Config{}
	err = yaml.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Panicf("Error (%v)\n", err)
	}
	log.Printf("--- t:\n%v\n\n", config)

	//
	ctx, err := torigoya.InitContext(cwd)
	if err != nil {
		log.Panicf(err.Error())
	}

	//
	log.Printf("Server initializing...\n")
	e := make(chan error)
	go func() {
		if err := <- e; err != nil {
			log.Panicf("Error (%v)\n", err)
		}
		log.Printf("Server starts!\n")
	}()

	//
	torigoya.RunServer(":12321", ctx, e)
}
