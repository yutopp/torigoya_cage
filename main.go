package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"gopkg.in/yaml.v1"

	"./torigoya/cage"
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
	fmt.Printf( "%s\n", config_bytes )
	println("====")

	config := Config{}
	err = yaml.Unmarshal(config_bytes, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", config)

	println("hello!", torigoya.F())
}
