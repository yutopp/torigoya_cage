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

	"yutopp/cage"
	"path/filepath"
	"gopkg.in/v1/yaml"
)


// replace string "${base}"
var base_reg = regexp.MustCompile("\\$\\{base\\}")

type Config map[string]*struct {
	Host						string `yaml:"host"`
	Port						int `yaml:"port"`
	HostUser					string `yaml:"host_user"`
	LangProcConfigDir			string `yaml:"lang_proc_config_dir"`
	LangProcUpdateZipAddress	string `yaml:"lang_proc_update_zip_address"`

	ProcPackageType				string `yaml:"proc_package_type"`
	ProcPackageDebSourceList	string `yaml:"proc_package_deb_source_list"`
	IsDebugMode					bool `yaml:"is_debug_mode"`
}

//
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicf("Error (%v)\n", err)
	}

	log.Printf("Current working dir: %s\n", cwd)

	//
	config_path := flag.String("config_path", "config.yml", "path to config.yml")
	mode := flag.String("mode", "release", "select mode from config")
	update := flag.Bool("update", false, "do update")
	pid := flag.Int("pid", -1, "a pid that receives a signal when the server startup")
	flag.Parse()

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
		fmt.Printf("the mode \"%s\" is not seletable. choose from below\n", *mode)
		for k, _ := range config {
			fmt.Printf("-> %s\n", k)
		}
		os.Exit(-1)
	}

	// show
	log.Printf("Mode:               %s\n", *mode)
	log.Printf("Host:               %s\n", target_config.Host)
    log.Printf("Port:               %d\n", target_config.Port)
    log.Printf("HostUser:           %s\n", target_config.HostUser)
    log.Printf("Profiles:           %s\n", target_config.LangProcConfigDir)
    log.Printf("ProcZipAddress:     %s\n", target_config.LangProcUpdateZipAddress)
	log.Printf("ProcPackageType:    %s\n", target_config.ProcPackageType)

	var updater torigoya.PackageUpdater = nil
	switch target_config.ProcPackageType {
	case "deb":
		updater = &torigoya.DebPackageUpdater{
			SourceListPath: target_config.ProcPackageDebSourceList,
		}
	default:
		log.Panicf("ProcPackageType (%v) is not supported\n", target_config.ProcPackageType)
	}

	//
	// make context!
	ctx_opt := &torigoya.ContextOptions{
		BasePath: cwd,
		UserFilesBasePath: filepath.Join(cwd, "fugafugafuga"),

		SandboxExec: &torigoya.AwahoSandboxExecutor{
		},

		ProcConfigPath: target_config.LangProcConfigDir,
		ProcSrcZipAddress: target_config.LangProcUpdateZipAddress,
		PackageUpdater: updater,
	}
	ctx, err := torigoya.InitContext(ctx_opt)
	if err != nil {
		log.Panicf(err.Error())
	}

	if !ctx.HasProcTable() {
		log.Printf("Try to download/reload proc_table...\n")
		if err := ctx.UpdateProcTable(); err != nil {
			log.Panicf(err.Error())
		}
		if !ctx.HasProcTable() {
			log.Panicf("unexpected: assertion failed...")
		}
		log.Printf("Finished to download/reload proc_table...\n")
	}

	if *update {
		log.Printf("Update environment...\n")

		log.Printf("(1/3) Update packages...\n")
		if err := ctx.UpdatePackages(); err != nil {
			log.Panicf(err.Error())
		}

		log.Printf("(2/3) Update tables...\n")
		if err := ctx.UpdateProcTable(); err != nil {
			log.Panicf(err.Error())
		}

		log.Printf("(3/3) Complete!\n")
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
	torigoya.RunServer(target_config.Host, target_config.Port, ctx, e, *pid)
}
