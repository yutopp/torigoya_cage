//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import (
	"gopkg.in/v1/yaml"
)

type SelectableCommand struct {
	Default		[]string
	Select		[]string `yaml:"select,flow"`
}

type PhaseDetail struct {
	File					string
	Extension				string
	Command					string
	Env						map[string]string
	AllowedCommandLine		map[string]SelectableCommand `yaml:"allowed_command_line"`
	FixedCommandLine		[][]string `yaml:"fixed_command_line"`

}

type ProcProfile struct {
	Version						string
	IsBuildRequired				bool `yaml:"is_build_required"`
	IsLinkIndependent			bool `yaml:"is_link_independent"`

	Source, Compile, Link, Run	PhaseDetail
}

func MakeProcProfile(buffer []byte) (*ProcProfile, error){
	profile := &ProcProfile{}

	if err := yaml.Unmarshal(buffer, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
