//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

package torigoya

import(
	"log"
)


func delegateSandboxContainer( cloner_path string ) {
	log.Println("called SekiseiRunnerNodeServer::delegate_sandbox_container")

	/*
      @logger.debug ""

      #
      spawner_path = "#{File.dirname( __FILE__ )}/container_spawner/spawner"
      callback_path = Pathname.new( "#{File.dirname( __FILE__ )}/callback.rb" ).relative_path_from( Pathname.new( Dir.pwd ) )

      @logger.debug "jhjhjhjhj #{spawner_path}"


      # create pipe
      read, write = IO.pipe

      #
      parameter_string = Base64.strict_encode64( Marshal.dump( *args ) )

      # create pipe
      read, write = IO.pipe

      #
      pid = Process.spawn( "#{spawner_path} '#{callback_path}' #{parameter_string}", { Config::Fd => write } )
      th = Process.detach( pid )

      #
      write.close
      result = read.read  # will blocking

      return_value = ( unless result.empty? then Marshal.load( result ) else nil end )
      value_processor.call( return_value ) if value_processor && !return_value.nil?

      # pid, detach_thread, result
      return pid, th, return_value

    end
*/
}
