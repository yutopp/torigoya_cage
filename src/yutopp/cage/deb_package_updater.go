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
	"os"
	"os/exec"
	"log"
)


type DebPackageUpdater struct {
	SourceListPath		string
	PackagePrefix		string
	InstallPrefix		string
}

func (u *DebPackageUpdater) Update() error {
	//
	if !fileExists(u.InstallPrefix) {
		if err := os.MkdirAll(u.InstallPrefix, 0755); err != nil {
			return err
		}
	}

	//
	regexPattern := fmt.Sprintf(`^(%s[^ ]+)( - )(.*)`, u.PackagePrefix)
	log.Printf("Regex pattern : %s\n", regexPattern)
	torigoyaDebPackageMatcher := regexp.MustCompile(regexPattern)

	// update packages for torigoya
	out, err := exec.Command("sudo", "apt-get", "update", "-o", "Dir::Etc::sourcelist=", u.SourceListPath, "-o", "Dir::Etc::sourceparts=", "-", "-o", "APT::Get::List-Cleanup=", "0").CombinedOutput()
	log.Printf("DebPackageUpdater apt-get update : %s\n", out)
	if err != nil {
		return errors.New("DebPackageUpdater error: " + err.Error())
	}

	// search packages for torigoya
	packagePattern := fmt.Sprintf("%s*", u.PackagePrefix)
	log.Printf("Search packages : %s\n", packagePattern)
	out, err = exec.Command("apt-cache", "search", packagePattern).Output()
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
	for i, p := range formatted_packages {
		log.Printf("(%d/%d) %s\n", i+1, len(formatted_packages), p)

		cmd := exec.Command("sudo", "apt-get", "install", "-y", "--force-yes", p)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			m := fmt.Sprintf("DebPackageUpdater error: on installing [%s]\n", p)
			log.Print(m)
			return errors.New(m)
		}
	}

	log.Printf("DebPackageUpdater info: try to upgrade: %v\n", formatted_packages)
	if out, err := exec.Command("sudo", append([]string{"apt-get", "upgrade", "-y", "--force-yes"}, formatted_packages...)...).CombinedOutput(); err != nil {
		m := fmt.Sprintf("DebPackageUpdater error: on upgrading [%s]\n", out)
		log.Print(m)
		return errors.New(m)
	}

	return nil
}

func (u *DebPackageUpdater) GetInstallPrefix() string {
	return u.InstallPrefix
}
