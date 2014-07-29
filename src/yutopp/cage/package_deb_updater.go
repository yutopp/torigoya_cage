//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

// +build linux

package torigoya

import(
	"fmt"
	"strings"
	"regexp"
	"errors"
	"os/exec"
	"log"
)


var torigoyaDebPackageMatcher = regexp.MustCompile(`^(torigoya-[^ ]+)( - )(.*)`)

type DebPackageUpdater struct {
	SourceListPath		string
}

func (u *DebPackageUpdater) Update() error {
	// update packages for torigoya
	out, err := exec.Command("sudo", "apt-get", "update", "-o", "Dir::Etc::sourcelist=", u.SourceListPath, "-o", "Dir::Etc::sourceparts=", "-", "-o", "APT::Get::List-Cleanup=", "0").CombinedOutput()
	log.Printf("DebPackageUpdater apt-get update : %s\n", out)
	if err != nil {
		return errors.New("DebPackageUpdater error: " + err.Error())
	}

	// search packages for torigoya
	out, err = exec.Command("apt-cache", "search", "torigoya-*").Output()
	if err != nil {
		return errors.New("DebPackageUpdater error: Couldn't search packages for torigoya")
	}

	//
	packages := strings.Split(string(out), "\n")
	var matched_packages []string = []string{}
	for _, p := range packages {
		if torigoyaDebPackageMatcher.MatchString(p) {
			matched_packages = append(matched_packages, p)
		}
	}

	if len(matched_packages) == 0 {
		log.Printf("DebPackageUpdater info: There are no packages to update\n")
		return nil
	}

	formatted_packages := []string{}
	for _, s := range packages {
		matched := torigoyaDebPackageMatcher.FindStringSubmatch(s)
		if len(matched) < 1 { continue }
		formatted_packages = append(formatted_packages, matched[1])
	}

	log.Printf("DebPackageUpdater info: try to install: %v\n", formatted_packages)
	if out, err := exec.Command("sudo", append([]string{"apt-get", "install", "-y", "--force-yes"}, formatted_packages...)...).CombinedOutput(); err != nil {
		m := fmt.Sprintf("DebPackageUpdater error: on installing [%s]\n", out)
		log.Print(m)
		return errors.New(m)
	}

	log.Printf("DebPackageUpdater info: try to upgrade: %v\n", formatted_packages)
	if out, err := exec.Command("sudo", append([]string{"apt-get", "upgrade", "-y", "--force-yes"}, formatted_packages...)...).CombinedOutput(); err != nil {
		m := fmt.Sprintf("DebPackageUpdater error: on upgrading [%s]\n", out)
		log.Print(m)
		return errors.New(m)
	}

	return nil
}
