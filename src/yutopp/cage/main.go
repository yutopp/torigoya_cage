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
	"regexp"
	"flag"
	"fmt"
	"os"
	"io/ioutil"

	"yutopp/torigoya/cage"

	"gopkg.in/v1/yaml"
)

// replace string "${base}"
var base_reg = regexp.MustCompile("\\$\\{base\\}")

type Config map[string]*struct {
	Host				string `yaml:"host"`
	Port				int `yaml:"port"`
	HostUser			string `yaml:"host_user"`
	LangProcConfigDir	string `yaml:"lang_proc_config_dir"`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicf("Error (%v)\n", err)
	}

	log.Printf("%s\n", cwd)

	//
	config_path := flag.String("config_path", "config.yml", "path to config.yml")
	mode := flag.String("mode", "local_debug", "select mode from config")

	//
	config_bytes, err := ioutil.ReadFile(*config_path)
	if err != nil {
		log.Panicf("There is no \"%s\" file...", *config_path)
	}

	//
	config := Config{}
	err = yaml.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Panicf("Error (%v)\n", err)
	}
	for _, v := range config {
		// replace meta string to instance
		v.LangProcConfigDir = base_reg.ReplaceAllString(v.LangProcConfigDir, cwd)
	}

	//
	target_config, ok := config[*mode]
	if !ok {
		fmt.Printf("the mode \"%s\" is not seletable. choose from below", *mode)
		for k, _ := range config {
			fmt.Printf("-> %s", k)
		}
		os.Exit(-1)
	}

	// show
	log.Printf("Host:       %s\n", target_config.Host)
	log.Printf("Port:       %d\n", target_config.Port)
	log.Printf("HostUser:   %s\n", target_config.HostUser)
	log.Printf("Profiles:   %s\n", target_config.LangProcConfigDir)

	//
	ctx, err := torigoya.InitContext(cwd, target_config.HostUser, target_config.LangProcConfigDir)
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

	// host, port
	torigoya.RunServer(target_config.Host, target_config.Port, ctx, e)
}
