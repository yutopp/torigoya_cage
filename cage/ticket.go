//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

//
// Ticket[1]
//   SourceCodes[N]
//   BuildInst[1]
//     CompileOption[1]
//     LinkOption[1]
//   RunInst[1]
//     Input[1]
//		 Stdin[1]
//       ExecOption[1]

// ========================================
// For source codes, stdins
type SourceData struct {
	Name         string `codec:"name"`
	Data         []byte `codec:"data"`
	IsCompressed bool   `codec:"is_compressed"`
}

func (s *SourceData) convertToTextContent() (*TextContent, error) {
	data, err := func() ([]byte, error) {
		// TODO: check that data is compressed
		return s.Data, nil
	}()
	if err != nil {
		return nil, err
	}

	return &TextContent{
		Name: s.Name,
		Data: data,
	}, nil
}

func convertSourcesToContents(
	sources []*SourceData,
) (source_contents []*TextContent, err error) {
	source_contents = make([]*TextContent, len(sources))

	for i, s := range sources {
		// collect file names
		source_contents[i], err = s.convertToTextContent()
		if err != nil {
			return nil, err
		}
	}

	return source_contents, nil
}

// ========================================
type ExecutionSetting struct {
	Command          string   `codec:"command"`
	Envs             []string `codec:"envs"`
	CpuTimeLimit     uint64   `codec:"cpu_time_limit"`
	MemoryBytesLimit uint64   `codec:"memory_bytes_limit"`
}

// ========================================
type BuildInstruction struct {
	CompileSetting *ExecutionSetting `codec:"compile_setting"`
	LinkSetting    *ExecutionSetting `codec:"link_setting,omitempty"`
}

func (build_inst *BuildInstruction) IsLinkIndependent() bool {
	return build_inst.LinkSetting != nil
}

// ========================================
type RunInstruction struct {
	Stdin      *SourceData       `codec:"stdin,omitempty"`
	RunSetting *ExecutionSetting `codec:"run_setting"`
}

// ========================================
type ExecutionSpec struct {
	BuildInst *BuildInstruction `codec:"build_inst,omitempty"`
	RunInsts  []*RunInstruction `codec:"run_insts,omitempty"`
}

func (spec *ExecutionSpec) IsBuildRequired() bool {
	return spec.BuildInst != nil
}

// ========================================
type Ticket struct {
	BaseName  string           `codec:"base_name"`
	Sources   []*SourceData    `codec:"sources"`
	ExecSpecs []*ExecutionSpec `codec:"exec_specs,omitempty"`
}
