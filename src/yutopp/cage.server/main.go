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
	"flag"
	"fmt"
	"os"
	"io/ioutil"

	"yutopp/cage"
	"gopkg.in/v1/yaml"
)

type updaterConfig struct {
	Type			string	`yaml:"type"`
	DebSourceList	string	`yaml:"source_list"`
	PackagePrefix	string	`yaml:"package_prefix"`
	InstallPrefix	string	`yaml:"install_prefix"`
}

func (c *updaterConfig) String() string {
	return "Type=" + c.Type + " / DebSourceList=" + c.DebSourceList
}

type sandboxConfig struct {
	Type			string	`yaml:"type"`
	AwahoExecutable	string	`yaml:"executable_path"`
}

func (c *sandboxConfig) String() string {
	return "Type=" + c.Type + " / AwahoExecutable=" + c.AwahoExecutable
}

type Config map[string]*struct {
	Host			string			`yaml:"host"`
	Port			int				`yaml:"port"`
	Updater			*updaterConfig	`yaml:"updater,omitempty"`
	SandboxExecutor	*sandboxConfig	`yaml:"sandbox"`
	IsDebugMode		bool			`yaml:"is_debug_mode"`
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
		log.Panicf("Loading Config: Error (%v)\n", err)
	}

	//
	target_config, ok := config[*mode]
	if !ok {
		fmt.Printf("The mode \"%s\" is not seletable. choose from below\n", *mode)
		for k, _ := range config {
			fmt.Printf("-> %s\n", k)
		}
		os.Exit(-1)
	}

	var updater torigoya.PackageUpdater = nil
	if target_config.Updater != nil {
		switch target_config.Updater.Type {
		case "deb":
			updater = &torigoya.DebPackageUpdater{
				SourceListPath: target_config.Updater.DebSourceList,
				PackagePrefix: target_config.Updater.PackagePrefix,
				InstallPrefix: target_config.Updater.InstallPrefix,
			}

		default:
			log.Panicf(
				"UpdaterType (%v) is not supported\n", target_config.Updater.Type,
			)
		}
	}


	if target_config.SandboxExecutor == nil {
		log.Panicf("sandbox option is required")
	}
	var sandbox torigoya.SandboxExecutor = nil
	switch target_config.SandboxExecutor.Type {
		case "awaho":
		sandbox, err = torigoya.MakeAwahoSandboxExecutor(
			target_config.SandboxExecutor.AwahoExecutable,
		)
		if err != nil {
			log.Panicf(err.Error())
		}

	default:
		log.Panicf(
			"SandboxType (%v) is not supported\n", target_config.SandboxExecutor.Type,
		)
	}

	// show
	log.Printf("==== Config ====")
	log.Printf("Mode    :  %s", *mode)
	log.Printf("Host    :  %s", target_config.Host)
    log.Printf("Port    :  %d", target_config.Port)
	if target_config.Updater != nil {
		log.Printf("Updater :  %s", target_config.Updater)
	}
	log.Printf("Sandbox :  %s", target_config.SandboxExecutor)


	// make context!
	ctx_opt := &torigoya.ContextOptions{
		BasePath: cwd,
		UserFilesBasePath: "/tmp/cage_recieved_files",
		PackageInstalledBasePath: target_config.Updater.InstallPrefix,

		SandboxExec: sandbox,
		PackageUpdater: updater,
	}
	ctx, err := torigoya.InitContext(ctx_opt)
	if err != nil {
		log.Panicf(err.Error())
	}

	if *update {
		log.Printf("Update environment...\n")

		log.Printf("(1/2) Update packages...\n")
		if err := ctx.UpdatePackages(); err != nil {
			log.Panicf(err.Error())
		}

		log.Printf("(2/2) Complete!\n")
	}

	//
	log.Printf("Server initializing...\n")
	e := make(chan error)
	go func() {
		if err := <- e; err != nil {
			log.Panicf("Server error: %v\n", err)
		}
		log.Printf("Server starts!\n")
	}()

	// host, port
	torigoya.RunServer(target_config.Host, target_config.Port, ctx, e)
}
