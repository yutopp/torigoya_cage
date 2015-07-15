//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya


// ========================================
// For source codes, stdins
type SourceData struct {
	Name			string	`codec:"name"`
	Data			[]byte	`codec:"data"`
	IsCompressed	bool	`codec:"is_compressed"`
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


// ========================================
type ExecutionSetting struct {
	Args				[]string	`codec:"args"`
	Envs				[]string	`codec:"envs"`
	CpuTimeLimit		uint64		`codec:"cpu_time_limit"`
	MemoryBytesLimit	uint64		`codec:"memory_bytes_limit"`
}


// ========================================
type BuildInstruction struct {
	CompileSetting		*ExecutionSetting	`codec:"compile_setting"`
	LinkSetting			*ExecutionSetting	`codec:"link_setting,omitempty"`
}

func (build_inst *BuildInstruction) IsLinkIndependent() bool {
	return build_inst.LinkSetting != nil
}


// ========================================
type Input struct{
	Stdin				*SourceData			`codec:"stdin,omitempty"`
	RunSetting			*ExecutionSetting	`codec:"run_setting"`
}


// ========================================
type RunInstruction struct {
	Inputs				[]Input			`codec:"inputs"`
}


// ========================================
type Ticket struct {
	BaseName		string				`codec:"base_name"`
	Sources			[]*SourceData		`codec:"sources"`
	BuildInst		*BuildInstruction	`codec:"build_inst,omitempty"`
	RunInst			*RunInstruction		`codec:"run_inst,omitempty"`
}

func (ticket *Ticket) IsBuildRequired() bool {
	return ticket.BuildInst != nil
}
