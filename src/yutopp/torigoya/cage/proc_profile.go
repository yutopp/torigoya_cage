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
	"os"
	"io"
	"io/ioutil"
	"path/filepath"
	"net/http"

	"gopkg.in/v1/yaml"
	"github.com/mattn/go-shellwords"
)


// ==================================================
type SelectableCommand struct {
	Default		[]string
	Select		[]string `yaml:"select,flow"`
}

func (sc *SelectableCommand) IsEmpty() bool { return sc.Default == nil || sc.Select == nil }


type PhaseDetail struct {
	File					string
	Extension				string
	Command					string
	Env						map[string]string
	AllowedCommandLine		map[string]SelectableCommand `yaml:"allowed_command_line"`
	FixedCommandLine		[][]string `yaml:"fixed_command_line"`
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
		args = append(args, v...)
	}

	// fixed commands
	for _, v := range pd.FixedCommandLine {
		args = append(args, v...)
	}

	// user command
	u_args, err := shellwords.Parse(command_line)
	if err != nil {
		return nil, err
	}
	args = append(args, u_args...)

	return args, nil
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
	IsBuildRequired				bool `yaml:"is_build_required"`
	IsLinkIndependent			bool `yaml:"is_link_independent"`

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
func makeProcProfileFromBuf(buffer []byte) (ProcProfile, error) {
	profile := ProcProfile{}

	if err := yaml.Unmarshal(buffer, &profile); err != nil {
		return profile, err
	}

	return profile, nil
}

func makeProcProfileFromPath(filepath string) (ProcProfile, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return ProcProfile{}, err
	}

	return makeProcProfileFromBuf(b)
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

	return makeProcDescriptionListFromBuf(b)
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

	proc_profile, ok := proc_unit.Versioned[proc_version]
	if !ok {
		return nil, errors.New("This proc_version is not registerd")
	}

	return &proc_profile, nil
}


func (pt *ProcConfigTable) UpdateFromWeb(address string) error {
	tmp_file, err := ioutil.TempFile("", "torigoya_tmp_")
	if err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}

	response, err := http.Get(address)
	if err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}
	defer response.Body.Close()

	if _, err := io.Copy(tmp_file, response.Body); err != nil {
		return errors.New("ProcConfigTable.ProcConfigTable error: " + err.Error())
	}

	return nil
}
