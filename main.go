//
// Copyright yutopp 2014 - 2016.
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/v1/yaml"
	"torigoya_cage/cage"
)

///
type sandboxConfig struct {
	Type                string `yaml:"type"`
	AwahoExecutablePath string `yaml:"executable_path"`
	HostMountDir        string `yaml:"host_mount_dir"`
	GuestMountDir       string `yaml:"guest_mount_dir"`
}

func (c *sandboxConfig) String() string {
	return "Type = " + c.Type + " / Path = " + c.AwahoExecutablePath
}

//
type Config map[string]*struct {
	Host            string         `yaml:"host"`
	Port            int            `yaml:"port"`
	SandboxExecutor *sandboxConfig `yaml:"sandbox"`
	IsDebugMode     bool           `yaml:"is_debug_mode"`
	ExecDir         string         `yaml:"exec_dir"`
}

//
func main() {
	config_path := flag.String("config", "config.yml", "path to config.yml")
	mode := flag.String("mode", "release", "select mode from config")
	flag.Parse()

	config_full_path, err := filepath.Abs(*config_path)
	if err != nil {
		log.Panicf("Failed to get abs path of config: %v", err)
	}
	config_dir := filepath.Dir(config_full_path)

	//
	config_bytes, err := ioutil.ReadFile(*config_path)
	if err != nil {
		log.Panicf("Failed to read config \"%s\": %v", *config_path, err)
	}

	//
	config := Config{}
	err = yaml.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Panicf("Failed to unmarshal config: %v\n", err)
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

	if target_config.SandboxExecutor == nil {
		log.Panicf("sandbox option is required")
	}
	var sandbox torigoya.SandboxExecutor = nil
	switch target_config.SandboxExecutor.Type {
	case "awaho":
		sandbox, err = torigoya.MakeAwahoSandboxExecutor(
			config_dir,
			target_config.SandboxExecutor.AwahoExecutablePath,
			target_config.SandboxExecutor.HostMountDir,
			target_config.SandboxExecutor.GuestMountDir,
		)
		if err != nil {
			log.Panicf(err.Error())
		}

	default:
		log.Panicf(
			"SandboxType (%v) is not supported\n", target_config.SandboxExecutor.Type,
		)
	}

	//
	if target_config.ExecDir == "" {
		log.Panicf("ExecDir is empty", target_config.ExecDir)
	}
	execDir := torigoya.NormalizePath(config_dir, target_config.ExecDir)

	// show
	log.Printf("==== Config ====")
	log.Printf("ConfigDir :  %s", config_dir)
	log.Printf("Mode      :  %s", *mode)
	log.Printf("Host      :  %s", target_config.Host)
	log.Printf("Port      :  %d", target_config.Port)
	log.Printf("Sandbox   :  %s", target_config.SandboxExecutor)
	log.Printf("ExecDir   :  %s", execDir)

	// make context!
	ctx_opt := &torigoya.ContextOptions{
		BasePath:          execDir,
		UserFilesBasePath: "/tmp/cage_received_files",
		SandboxExec:       sandbox,
	}
	ctx, err := torigoya.InitContext(ctx_opt)
	if err != nil {
		log.Panicf(err.Error())
	}

	//
	log.Printf("Server initializing...\n")
	e := make(chan error)
	go func() {
		if err := <-e; err != nil {
			log.Panicf("Server error: %v\n", err)
		}
		log.Printf("Server starts!\n")
	}()

	// host, port
	torigoya.RunServer(target_config.Host, target_config.Port, ctx, e)
}
