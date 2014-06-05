// +build linux

package torigoya

import(
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"path/filepath"
)

type Context struct {
	hostUser		*user.User
	ticketDir		string
}


func InitContext() *Context {
	//
	ticket_dir := "/tmp/ticket"

	//
	host_user_name := "yutopp"
	host_user, err := user.Lookup(host_user_name)
	if err != nil {
	}

	//
	//
	if !fileExists(ticket_dir) {
		err := os.Mkdir(ticket_dir, os.ModeDir | 0700)
		if err != nil {
			panic(fmt.Sprintf("Couldn't create directory %s", ticket_dir))
		}
	}

	return &Context{
		hostUser: host_user,
		ticketDir: ticket_dir,
	}
}

func F() int {
	ctx := InitContext()

	ctx.createTarget("aaa")

	return 42
}

func expectRoot() {
	if os.Geteuid() != 0 {
		panic("run this program as root")
	}
}

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

//
func (ctx *Context) createTarget(code_name string/*, revision, group_id, open_type, &block */) {
	log.Println("called SekiseiRunnerNodeServer::create_target")

    // expectRoot()

	fmt.Printf( "Euid -> %d\n", os.Geteuid() )

	// In posix, Uid only contains numbers
	host_user_id, _ := strconv.Atoi(ctx.hostUser.Uid)

	if !fileExists(ctx.ticketDir) {
		panic(fmt.Sprintf("directory %s is not existed", ctx.ticketDir))
	}

	//
	user_dir_path := filepath.Join(ctx.ticketDir, code_name)

	//
	if fileExists(user_dir_path) {
		err := os.RemoveAll(user_dir_path)
		if err != nil {
			panic(fmt.Sprintf("Couldn't remove directory %s", user_dir_path))
		}
	}

	//
	err := os.Mkdir(user_dir_path, os.ModeDir | 0700)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create directory %s", user_dir_path))
	}

	fmt.Printf("host uid: %s\n", ctx.hostUser.Uid)

	// TODO: fix it
	managed_group_id := 1000
	source_filename := "prog.cpp"

	err = os.Chown(user_dir_path, host_user_id, managed_group_id)
	if err != nil {
	}
	err = os.Chmod(user_dir_path, 0777)
	if err != nil {
	}



	f, err := os.OpenFile(source_filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic("")
	}
	defer f.Close()


	err = os.Chown(source_filename, host_user_id, managed_group_id)
	if err != nil {
	}
	err = os.Chmod(source_filename, 0400)	// r--/---/---
	if err != nil {
	}

/*
      profile = runner.name_profile
      directory_name = "#{code_name}.#{revision}" # Sekisei.build_name( language_id, code_name, revision )

      #
      Dir.chdir( TicketsDir ) do
        FileUtils.remove_entry_secure( directory_name, true ) if File.exists?( directory_name )
        # make inputs source code data holder
        Dir.mkdir( directory_name )
        # delegate ownership to runner user
        File.chown( @host_user_id, group_id, directory_name )
        File.chmod( Utility.make_permission(:r,:w,:x), directory_name )

        # create source code
        Dir.chdir( directory_name ) do
          File.open( profile[:filename_source], open_type ) do |f|
            block.call( f )
          end
          File.chmod( Utility.make_permission(:r), profile[:filename_source] )
          File.chown( @host_user_id, group_id, profile[:filename_source] )
        end
      end
      return Pathname.new( File.expand_path( "#{TicketsDir}/#{directory_name}" ) )
    end
*/
}
