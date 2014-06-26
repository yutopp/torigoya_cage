//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package main

import (
	"fmt"
	"log"
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
	config_bytes, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic("there is no \"config.yml\" file...")
	}

	//
	config := Config{}
	err = yaml.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", config)

	//
	fmt.Printf("Server starts...!\n")
	if err := torigoya.RunServer(":12321", nil); err != nil {
		fmt.Printf("Error (%v)\n", err)
	}
}
