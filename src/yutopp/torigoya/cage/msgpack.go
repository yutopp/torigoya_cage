//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"errors"

	"github.com/ugorji/go/codec"
)


var msgPackHandler codec.MsgpackHandle

func readUInt(v interface{}) (uint64, bool) {
	switch v.(type) {
	case int64:
		return uint64(v.(int64)), true
	case uint64:
		return v.(uint64), true
	default:
		return 0, false
	}
}


// ========================================
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
func MakeExecutionSettingFromTuple(tupled interface{}) (*ExecutionSetting, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(total)") }
	if len(interface_array) != 4 { return nil, errors.New("ExecutionSetting::invalid data(num of lement)") }

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

	//
	cpu_time_limit, ok := readUInt(interface_array[2])
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(2)") }

	//
	memory_bytes_limit, ok := readUInt(interface_array[3])
	if !ok { return nil, errors.New("ExecutionSetting::invalid data(3)") }

	//
	return &ExecutionSetting{
		CommandLine: string(command_line_bytes),
		StructuredCommand: structured_commands,
		CpuTimeLimit: cpu_time_limit,
		MemoryBytesLimit: memory_bytes_limit,
	}, nil
}


// ========================================
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
func MakeInputFromTuple(tupled interface{}) (*Input, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("Input::invalid data(total)") }
	if len(interface_array) != 2 { return nil, errors.New("Input::invalid data(num of lement)") }

	//
	input, err := MakeSourceDataFromTuple(interface_array[0])
	if err != nil { return nil, errors.New("Input::invalid data(0)") }

	//
	run_setting, err := MakeExecutionSettingFromTuple(interface_array[1])
	if err != nil { return nil, errors.New("Input::invalid data(1)") }

	return &Input{
		input: input,
		setting: run_setting,
	}, nil
}


// ========================================
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
func MakeTicketFromTuple(tupled interface{}) (*Ticket, error) {
	if tupled == nil { return nil, nil }
	interface_array, ok := tupled.([]interface{})
	if !ok { return nil, errors.New("Ticket::invalid data(total)") }
	if len(interface_array) != 6 { return nil, errors.New("Ticket::invalid data(num of lement)") }

	//
	base_name_bytes, ok := interface_array[0].([]byte)
	if !ok { return nil, errors.New("Ticket::invalid data(0)") }

	//
	proc_id, ok := readUInt(interface_array[1])
	if !ok { return nil, errors.New("Ticket::invalid data(1)") }

	//
	proc_version_bytes, ok := interface_array[2].([]byte)
	if !ok { return nil, errors.New("Ticket::invalid data(2)") }

	//
	sources_interface_array, ok := interface_array[3].([]interface{})
	if !ok { return nil, errors.New("Ticket::invalid data(3)") }
	var sources []*SourceData
	for _, source_interface := range sources_interface_array {
		source, err := MakeSourceDataFromTuple(source_interface)
		if err != nil { return nil, err }

		sources = append(sources, source)
	}

	//
	bi, err := MakeBuildInstructionFromTuple(interface_array[4])
	if err != nil { return nil, err }

	//
	ri, err := MakeRunInstructionFromTuple(interface_array[5])
	if err != nil { return nil, err }

	//
	return &Ticket{
		BaseName: string(base_name_bytes),
		ProcId: proc_id,
		ProcVersion: string(proc_version_bytes),
		Sources: sources,
		BuildInst: bi,
		RunInst: ri,
	}, nil
}
