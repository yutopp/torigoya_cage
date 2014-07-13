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
	"log"
	"fmt"
	"errors"
	"path/filepath"
)


// send this message to sandbox process
type ExecMessage struct {
	Profile				*ProcProfile
	StdinFilePath		*string
	Setting				*ExecutionSetting
	Mode				int
}

const (
	CompileMode = iota
	LinkMode
	RunMode
)

var (
	compileFailedError	= errors.New("compile failed")
	linkFailedError		= errors.New("link failed")
	buildFailedError	= errors.New("build failed")
)

// ========================================
func (ctx *Context) ExecTicket(
	ticket				*Ticket,
	callback			invokeResultRecieverCallback,
) error {
	// lookup language proc profile
	proc_profile, err := ctx.procConfTable.Find(ticket.ProcId, ticket.ProcVersion)
	if err != nil {
		return err
	}

	//
	if err := ctx.execManagedBuild(proc_profile, ticket.BaseName, ticket.Sources, ticket.BuildInst, callback); err != nil {
		if err == buildFailedError {
			return nil
		} else {
			return err
		}
	}
	//

	// run
	if errs := ctx.execManagedRun(proc_profile,	ticket.BaseName, ticket.Sources, ticket.RunInst, callback); errs != nil {
		// TODO: proess error
		for err := range errs {
			fmt.Printf("??? %v\n", err)
		}
		return errors.New("ababa")
	}

	return nil
}


//
func (ctx *Context) execManagedBuild(
	proc_profile		*ProcProfile,
	base_name			string,
	sources				[]*SourceData,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	//
	user_dir_path := ctx.makeUserDirName(base_name)
	user_home_path := ctx.jailedUserDir
	bin_base_path := filepath.Join(ctx.basePath, "bin")

	//
	if proc_profile.IsBuildRequired {
		// build required processor
		if err := runAsManagedUser(func(jailed_user *JailedUserInfo) error {
			// compile phase
			// map files
			if err := ctx.mapSources(base_name, sources, jailed_user, proc_profile); err != nil {
				return err
			}

			//
			if err := ctx.invokeCompileCommand(user_dir_path, user_home_path, bin_base_path, jailed_user, proc_profile, base_name, sources, build_inst, callback); err != nil {
				if err == compileFailedError {
					return buildFailedError
				} else {
					return err
				}
			}

			// link phase :: if link command is separated, so call linking commands
			if proc_profile.IsLinkIndependent {
				if err := ctx.cleanupMountedFiles(base_name); err != nil {
					return err
				}

				if err := ctx.invokeLinkCommand(user_dir_path, user_home_path, bin_base_path, jailed_user, proc_profile, base_name, sources, build_inst, callback); err != nil {
					if err == linkFailedError {
						return buildFailedError
					} else {
						return err
					}
				}
			}

			return nil

		}); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) execManagedRun(
	proc_profile		*ProcProfile,
	base_name			string,
	sources				[]*SourceData,
	run_inst			*RunInstruction,
	callback			invokeResultRecieverCallback,
) []error {
	log.Println(">> called invokeRunCommand")

	//
	user_dir_path := ctx.makeUserDirName(base_name)
	user_home_path := ctx.jailedUserDir
	bin_base_path := filepath.Join(ctx.basePath, "bin")

	// if it is build NOT required processor, sources have not been mapped yet
	if !proc_profile.IsBuildRequired {
		if err := runAsManagedUser(func(jailed_user *JailedUserInfo) error {
			// map files
			if err := ctx.mapSources(base_name, sources, jailed_user, proc_profile); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return []error{err}
		}
	}

	//
	var errs []error = nil
	// ========================================
	for index, input := range run_inst.Inputs {
		// TODO: async
		err := runAsManagedUser(func(jailed_user *JailedUserInfo) error {
			return ctx.invokeRunCommand(
				user_dir_path,
				user_home_path,
				bin_base_path,
				jailed_user,

				base_name,
				proc_profile,
				index,
				&input,
				callback,
			)
		})

		if err != nil {
			if errs == nil { errs = make([]error, 0) }
			errs = append(errs, err)
		}
	}

	return errs
}


// ========================================
// ========================================


func (ctx *Context) invokeCompileCommand(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,
	proc_profile		*ProcProfile,
	base_name			string,
	sources				[]*SourceData,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeCompileCommand")

	if build_inst == nil { return errors.New("compile_dataset is nil") }
	if build_inst.CompileSetting == nil {
		return errors.New("build_inst.CompileSetting is nil")
	}

	//
	message := BridgeMessage{
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		JailedUser: jailed_user,
		Message: ExecMessage{
			Profile: proc_profile,
			Setting: build_inst.CompileSetting,
			Mode: CompileMode,
		},
		IsReboot: false,
	}

	//
	build_output_stream := make(chan StreamOutput)
	go sendOutputToCallback(callback, build_output_stream, CompileMode, 0)

	//
	result, err := message.invokeProcessCloner(bin_base_path, build_output_stream)

	//
	close(build_output_stream)
	if err != nil { return err }
	sendResultToCallback(callback, result, CompileMode, 0)

	if result.IsFailed() {
		return compileFailedError
	}

	return nil
}

func (ctx *Context) invokeLinkCommand(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,
	proc_profile		*ProcProfile,
	base_name			string,
	sources				[]*SourceData,
	build_inst			*BuildInstruction,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeLinkCommand")

	//
	if build_inst == nil { return errors.New("compile_dataset is nil") }
	if build_inst.LinkSetting == nil {
		return errors.New("build_inst.LinkSetting is nil")
	}

	//
	message := BridgeMessage{
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		JailedUser: jailed_user,
		Message: ExecMessage{
			Profile: proc_profile,
			Setting: build_inst.LinkSetting,
			Mode: LinkMode,
		},
		IsReboot: false,
	}

	//
	link_output_stream := make(chan StreamOutput)
	go sendOutputToCallback(callback, link_output_stream, LinkMode, 0)

	//
	result, err := message.invokeProcessCloner(bin_base_path, link_output_stream)

	//
	close(link_output_stream)
	if err != nil { return err }
	sendResultToCallback(callback, result, LinkMode, 0)

	if result.IsFailed() {
		return linkFailedError
	}

	return nil
}

func (ctx *Context) invokeRunCommand(
	user_dir_path		string,
	user_home_path		string,
	bin_base_path		string,
	jailed_user			*JailedUserInfo,

	base_name			string,
	proc_profile		*ProcProfile,
	index				int,
	input				*Input,
	callback			invokeResultRecieverCallback,
) error {
	log.Println(">> called invokeRunInputCommand")

	// TODO: add lock
	// reassign base files to new user
	user_dir_path, input_path, err := ctx.reassignTarget(
		base_name,
		jailed_user.UserId,
		jailed_user.GroupId,
		func(base_directory_name string) (*string, error) {
			if input.stdin != nil {
				// stdin exists

				// unpack source codes
				stdin_content, err := convertSourceToContent(input.stdin)
				if err != nil { return nil, err }

				path, err := ctx.createInput(base_directory_name, jailed_user.GroupId, stdin_content)
				if err != nil { return nil, err }
				return &path, nil

			} else {
				// nothing to do
				return nil, nil
			}
		},
	)
	if err != nil { return err }

	//
	var stdin_path *string = nil
	if input.stdin != nil {
		if input_path == nil { return errors.New("invalid stdin file") }

		// adjust path to jailed env
		real_user_home_path := filepath.Join(user_dir_path, user_home_path)
		stdin_path_val, err := filepath.Rel(real_user_home_path, *input_path)
		if err != nil {}

		stdin_path = &stdin_path_val
	}

	//
	message := BridgeMessage{
		ChrootPath: user_dir_path,
		JailedUserHomePath: user_home_path,
		JailedUser: jailed_user,
		Message: ExecMessage{
			Profile: proc_profile,
			StdinFilePath: stdin_path,
			Setting: input.setting,
			Mode: RunMode,
		},
		IsReboot: false,
	}

	//
	run_output_stream := make(chan StreamOutput)
	go sendOutputToCallback(callback, run_output_stream, RunMode, index)

	//
	result, err := message.invokeProcessCloner(bin_base_path, run_output_stream)

	//
	close(run_output_stream)
	if err != nil { return err }
	sendResultToCallback(callback, result, RunMode, index)

	return nil
}


// ========================================
// ========================================


func (ctx *Context) mapSources(
	base_name			string,
	sources				[]*SourceData,
	jailed_user			*JailedUserInfo,
	proc_profile		*ProcProfile,
) error {
	// unpack source codes
	source_contents, err := convertSourcesToContents(sources)
	if err != nil {
		return err
	}

	//
	default_filename := fmt.Sprintf("%s.%s", proc_profile.Source.File, proc_profile.Source.Extension)
	fmt.Printf("DEFEDEDEAFAWF   %s", default_filename)

	//
	if _, err := ctx.createMultipleTargetsWithDefaultName(
		base_name,
		jailed_user.GroupId,
		source_contents,
		&default_filename,
	); err != nil {
		return errors.New("couldn't create multi target : " + err.Error());
	}

	return nil
}


// ========================================
// ========================================


//
type StreamOutputResult struct {
	Mode		int
	Index		int
	Output		*StreamOutput
}

func (r *StreamOutputResult) ToTuple() []interface{} {
	return []interface{}{ r.Mode, r.Index, r.Output.ToTuple() }
}


//
type StreamExecutedResult struct {
	Mode		int
	Index		int
	Result		*ExecutedResult
}
func (r *StreamExecutedResult) ToTuple() []interface{} {
	return []interface{}{ r.Mode, r.Index, r.Result.ToTuple() }
}


//
type invokeResultRecieverCallback		func(interface{})


//
func sendOutputToCallback(
	callback			invokeResultRecieverCallback,
	output_stream		chan StreamOutput,
	mode				int,
	index				int,
) {
	for out := range output_stream {
		if callback != nil {
			callback(&StreamOutputResult{
				Mode: mode,
				Index: index,
				Output: &out,
			})
		}
	}
}


//
func sendResultToCallback(
	callback			invokeResultRecieverCallback,
	result				*ExecutedResult,
	mode				int,
	index				int,
) {
	if callback != nil {
		callback(&StreamExecutedResult{
			Mode: mode,
			Index: index,
			Result: result,
		})
	}
}
