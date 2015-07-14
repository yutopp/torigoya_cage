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
)


// ========================================
// For source codes, stdins
type SourceData struct {
	Name			string
	Data			[]byte
	IsCompressed	bool
}

func convertSourcesToContents(
	sources []*SourceData,
) (source_contents []*TextContent, err error) {
	source_contents = make([]*TextContent, len(sources))

	//
	for i, s := range sources {
		// collect file names
		source_contents[i], err = convertSourceToContent(s)
		if err != nil { return nil,err }
	}

	return source_contents, nil
}

func convertSourceToContent(
	source *SourceData,
) (*TextContent, error) {
	data, err := func() ([]byte, error) {
		// TODO: check that data is compressed
		return source.Data, nil
	}()
	if err != nil {
		return nil, err
	}

	return &TextContent{
		Name: source.Name,
		Data: data,
	}, nil
}


func MakeSourceDataFromTuple(tupled interface{}) (*SourceData, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("SourceData::invalid data(total)") }
	if len(interface_array) != 3 { return nil, errors.New("SourceData::invalid data(num of lement)") }

	name_bytes, ok := interface_array[0].([]byte)
	if !ok { return nil, errors.New("SourceData::invalid data(0)") }

	data_byte, ok := interface_array[1].([]byte)
	if !ok { return nil, errors.New("SourceData::invalid data(1)") }

	is_compressed, ok := interface_array[2].(bool)
	if !ok { return nil, errors.New("SourceData::invalid data(2)") }

	return &SourceData{
		Name: string(name_bytes),
		Data: data_byte,
		IsCompressed: is_compressed,
	}, nil
}


// ========================================
type ExecutionSetting struct {
	Args				[]string
	Envs				[]string
	CpuTimeLimit		uint64
	MemoryBytesLimit	uint64
}


func MakeExecutionSettingFromTuple(tupled interface{}) (*ExecutionSetting, error) {
	if tupled == nil { return nil, nil }

	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(total)") }
	if len(interface_array) != 4 { return nil, errors.New("ExecutionSetting::invalid data(num of lement)") }
/*
	//
	command_line_bytes, ok := interface_array[0].([]byte)
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(0)") }

	//
	structured_commands_array, ok := interface_array[1].([]interface{})
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(1)") }

	structured_commands := make([][]string, len(structured_commands_array))
	for i, structured_command_array := range structured_commands_array {
		strings_array, ok := structured_command_array.([]interface{})
		if !ok { return nil, errors.New("ExecutionSetting::invalid data(1) in") }

		commands := make([]string, len(strings_array))
		for j, string_array := range strings_array {
			string_bytes, ok := string_array.([]byte)
			if !ok { return nil, errors.New("ExecutionSetting::invalid data(0) in bytes") }

			commands[j] = string(string_bytes)
		}
		structured_commands[i] = commands
	}
*/
	//
	cpu_time_limit, ok := readUInt(interface_array[2])
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(2)") }

	//
	memory_bytes_limit, ok := readUInt(interface_array[3])
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(3)") }

	//
	return &ExecutionSetting{
		Args: nil,
		Envs: nil,
		CpuTimeLimit: cpu_time_limit,
		MemoryBytesLimit: memory_bytes_limit,
	}, nil
}


// ========================================
type BuildInstruction struct {
	CompileSetting		*ExecutionSetting
	LinkSetting			*ExecutionSetting
}

func (build_inst *BuildInstruction) IsLinkIndependent() bool {
	return build_inst.LinkSetting != nil
}


func MakeBuildInstructionFromTuple(tupled interface{}) (*BuildInstruction, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("BuildInstruction::invalid data(total)") }
	if len(interface_array) != 2 { return nil, errors.New("BuildInstruction::invalid data(num of lement)") }

	compile_setting, err := MakeExecutionSettingFromTuple(interface_array[0])
	if err != nil { return nil, errors.New("BuildInstruction::invalid data(0)") }

	link_setting, err := MakeExecutionSettingFromTuple(interface_array[1])
	if err != nil { return nil, errors.New("BuildInstruction::invalid data(1)") }

	return &BuildInstruction{
		CompileSetting: compile_setting,
		LinkSetting: link_setting,
	}, nil
}


// ========================================
type Input struct{
	stdin				*SourceData
	setting				*ExecutionSetting
}


func MakeInputFromTuple(tupled interface{}) (*Input, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("Input::invalid data(total)") }
	if len(interface_array) != 2 { return nil, errors.New("Input::invalid data(num of lement)") }

	//
	stdin, err := MakeSourceDataFromTuple(interface_array[0])
	if err != nil { return nil, errors.New("Input::invalid data(0)") }

	//
	run_setting, err := MakeExecutionSettingFromTuple(interface_array[1])
	if err != nil { return nil, errors.New("Input::invalid data(1)") }

	return &Input{
		stdin: stdin,
		setting: run_setting,
	}, nil
}


// ========================================
type RunInstruction struct {
	Inputs				[]Input
}


func MakeRunInstructionFromTuple(tupled interface{}) (*RunInstruction, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("RunInstruction::invalid data(total)") }
	if len(interface_array) != 1 { return nil, errors.New("RunInstruction::invalid data(num of lement)") }

	inputs_array, ok := interface_array[0].([]interface{})
	if !ok { return nil, errors.New("RunInstruction::invalid data(0)") }

	//
	inputs := make([]Input, len(inputs_array))
	for i, input_array := range inputs_array {
		input, err := MakeInputFromTuple(input_array)
		if err != nil { return nil, err }

		inputs[i] = *input
	}

	return &RunInstruction{
		Inputs: inputs,
	}, nil
}


// ========================================
type Ticket struct {
	BaseName		string
	Sources			[]*SourceData
	BuildInst		*BuildInstruction
	RunInst			*RunInstruction
}

func (ticket *Ticket) IsBuildRequired() bool {
	return ticket.BuildInst != nil
}


func MakeTicketFromTuple(tupled interface{}) (*Ticket, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("Ticket::invalid data: not an array") }
	if len(interface_array) != 4 { return nil, errors.New("Ticket::invalid data: num of lement") }

	//
	base_name_bytes, ok := interface_array[0].([]byte)
	if !ok { return nil, errors.New("Ticket::invalid data(0)") }

	//
	sources_interface_array, ok := interface_array[1].([]interface{})
	if !ok { return nil, errors.New("Ticket::invalid data: sources [index: 0]") }
	var sources []*SourceData
	for _, source_interface := range sources_interface_array {
		source, err := MakeSourceDataFromTuple(source_interface)
		if err != nil { return nil, err }

		sources = append(sources, source)
	}

	//
	bi, err := MakeBuildInstructionFromTuple(interface_array[2])
	if err != nil { return nil, err }

	//
	ri, err := MakeRunInstructionFromTuple(interface_array[3])
	if err != nil { return nil, err }

	//
	return &Ticket{
		BaseName: string(base_name_bytes),
		Sources: sources,
		BuildInst: bi,
		RunInst: ri,
	}, nil
}
