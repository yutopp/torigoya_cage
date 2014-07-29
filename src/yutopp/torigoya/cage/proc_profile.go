//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"errors"
	"fmt"
	"strings"
	"os"
	"os/exec"
	"io"
	"regexp"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"encoding/json"

	"gopkg.in/v1/yaml"
	"github.com/mattn/go-shellwords"
)


// ==================================================
var headVersionPattern = regexp.MustCompile("^HEAD-")
var devVersionPattern = regexp.MustCompile("^DEV-")
var stableVersionPattern = regexp.MustCompile("^STABLE-")
var specialRecuest = map[string]*regexp.Regexp{
	"!=head": headVersionPattern,
	"!=dev": devVersionPattern,
	"!=stable": stableVersionPattern,
}

// ==================================================
type SelectableCommand struct {
	Default		[]string
	Select		[]string `json:"select"`
}

func (sc *SelectableCommand) IsEmpty() bool { return sc.Default == nil || sc.Select == nil }


type PhaseDetail struct {
	File					string
	Extension				string
	Command					string
	Env						map[string]string
	AllowedCommandLine		map[string]SelectableCommand `json:"allowed_command_line"`
	FixedCommandLine		[][]string `json:"fixed_command_line"`
}

func (pd *PhaseDetail) MakeCompleteArgs(
	command_line string,
	selected_options [][]string,
) ([]string, error) {
	for _, v := range selected_options {
		if err := pd.isValidOption(v); err != nil {
			return nil, err
		}
	}

	args := []string{}

	// command
	if len(pd.Command) == 0 {
		return nil, errors.New("command can not be empty")
	}
	args = append(args, pd.Command)

	// selected user commands(structured)
	for _, v := range selected_options {
		args = append(args, argCat(v)...)
	}

	// fixed commands
	for _, v := range pd.FixedCommandLine {
		args = append(args, argCat(v)...)
	}

	// user command
	u_args, err := shellwords.Parse(command_line)
	if err != nil {
		return nil, err
	}
	args = append(args, u_args...)

	return args, nil
}

func argCat(v []string) []string {
	// TODO: check length of array
	if len(v) == 2 {
		k := v[0]
		if k[len(k)-1] == ' ' {
			return []string{strings.TrimSpace(k), v[1]}
		} else {
			return []string{k+v[1]}
		}
	} else {
		return v
	}
}

func (pd *PhaseDetail) isValidOption(selected_option []string) error {
	if !( len(selected_option) == 1 || len(selected_option) == 2 ) {
		return errors.New(fmt.Sprintf("isValidOption::length of the option should be 1 or 2 (but %d)", len(selected_option)))
	}

	if val,ok := pd.AllowedCommandLine[selected_option[0]]; ok {
		if len(selected_option) == 2 {
			// TODO: fix to bin search
			for _, v := range pd.AllowedCommandLine[selected_option[0]].Select {
				if v == selected_option[1] {
					return nil
				}
			}
			return errors.New(fmt.Sprintf("isValidOption::value(%s) was not found in key(%s)", selected_option[1], selected_option[0]))

		} else {
			// selected option is only key
			if val.IsEmpty() {
				return nil

			} else {
				return errors.New("isValidOption::nil value can not be selected")
			}
		}

	} else {
		return errors.New(fmt.Sprintf("isValidOption::key(%s) was not found", selected_option[0]))
	}
}


type ProcProfile struct {
	Version						string
	IsBuildRequired				bool `json:"is_build_required"`
	IsLinkIndependent			bool `json:"is_link_independent"`

	Source, Compile, Link, Run	PhaseDetail
}


// ==================================================
type ProcDescription struct {
	Id			uint64
	Name		string
	Runnable	bool
	Path		string
}

type ProcDescriptionList []ProcDescription


// ==================================================
type ProcConfigTable map[uint64]ProcConfigUnit		// proc_id:config_unit
type ProcConfigUnit struct {
	Description		ProcDescription
	Versioned		map[string]ProcProfile
}


// ==================================================
// ==================================================
func makeProcProfileFromBufAsJSON(buffer []byte) (ProcProfile, error) {
	profile := ProcProfile{}

	if err := json.Unmarshal(buffer, &profile); err != nil {
		return profile, err
	}

	return profile, nil
}

func makeProcProfileFromPath(filepath string) (ProcProfile, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return ProcProfile{}, err
	}

	pt, err := makeProcProfileFromBufAsJSON(b);
	if err != nil {
		return pt, errors.New(fmt.Sprintf("In [%s] : %v", filepath, err))
	}

	return pt, nil
}


func makeProcDescriptionListFromBuf(buffer []byte) (ProcDescriptionList, error) {
	var index_list ProcDescriptionList

	if err := yaml.Unmarshal(buffer, &index_list); err != nil {
		return nil, err
	}

	return index_list, nil
}

func makeProcDescriptionListFromPath(filepath string) (ProcDescriptionList, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	pdl, err := makeProcDescriptionListFromBuf(b)
	if err != nil {
		return pdl, errors.New(fmt.Sprintf("In [%s] : %v", filepath, err))
	}

	return pdl, nil
}


func globProfiles(proc_path string) (map[string]ProcProfile, error) {
	result := make(map[string]ProcProfile)

	if err := filepath.Walk(proc_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// only json files will be accepted
		if filepath.Ext(path) != ".json" {
			return nil
		}

		proc_profile, err := makeProcProfileFromPath(path)
		if err != nil {
			return err;
		}

		result[proc_profile.Version] = proc_profile

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}


func LoadProcConfigs(proc_prof_base_path string) (ProcConfigTable, error) {
	var result = make(ProcConfigTable)

	index_list, err := makeProcDescriptionListFromPath(filepath.Join(proc_prof_base_path, "languages.yml"))
	if err != nil {
		return nil, err
	}

	for _, proc_index := range index_list {
		versioned_proc_profiles, err := globProfiles(filepath.Join(proc_prof_base_path, proc_index.Path))
		if err != nil {
			return nil, err
		}

		result[proc_index.Id] = ProcConfigUnit{
			proc_index,
			versioned_proc_profiles,
		}
	}

	return result, nil
}

func (pt *ProcConfigTable) Find(proc_id uint64, proc_version string) (*ProcProfile, error) {
	proc_unit, ok := (*pt)[proc_id]
	if !ok {
		return nil, errors.New("This proc_id is not registerd")
	}

	if reg, ok := specialRecuest[proc_version]; ok {
		for k, v := range proc_unit.Versioned {
			if reg.MatchString(k) {
				return &v, nil
			}
		}
		return nil, errors.New("This proc_version(special) is not registerd")

	} else {
		proc_profile, ok := proc_unit.Versioned[proc_version]
		if !ok {
			return nil, errors.New("This proc_version is not registerd")
		}
		return &proc_profile, nil
	}
}


func (pt *ProcConfigTable) UpdateFromWeb(address string, base_path string) error {
	// make temporary file to save zip file
	tmp_file, err := ioutil.TempFile("", "torigoya_tmp_")
	if err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}

	// get data from web
	response, err := http.Get(address)
	if err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}
	defer response.Body.Close()

	// copy binary to the temporary file
	if _, err := io.Copy(tmp_file, response.Body); err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}

	// extract to...
	// -o: force extract
	// -d: target dir
	cmd := exec.Command("unzip", "-o", tmp_file.Name(), "-d", "files")

	return cmd.Run()
}
