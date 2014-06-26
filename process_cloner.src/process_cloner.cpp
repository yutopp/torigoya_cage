//
// Copyright yutopp 2014 - .
//
// Distributed under the Boost Software License, Version 1.0.
// (See accompanying file LICENSE_1_0.txt or copy at
// http://www.boost.org/LICENSE_1_0.txt)
//

#include <iostream>
#include <algorithm>
#include <string>
#include <memory>
#include <array>
#include <cstring>
#include <cassert>

#include <unistd.h>
#include <sys/types.h>
#include <sys/mount.h>
#include <sys/prctl.h>
#include <sys/wait.h>
#include <sched.h>
#include <signal.h>


static_assert( sizeof( pid_t ) == sizeof( int ), "sizeof( pid_t ) != sizeof( int )" );


//
int fork_shell( void* /* unused */ )
{
    // mount procfs
    //::mount( "procfs", "/proc", "proc", 0, nullptr );

    // change "callee" process name
    //::prctl( PR_SET_NAME, "torigoya bridge" );

    //
    char* const callback_executable_r = getenv( "callback_executable" );
    char* const packed_torigoya_content_r = getenv( "packed_torigoya_content" );
    if ( callback_executable_r == nullptr || packed_torigoya_content_r == nullptr ) {
        std::cerr << "A number of parameters is not enough." << std::endl;
        return -1;
    }
/*
    std::cout << "c = " << callback_executable_r << std::endl
              << "p = " << packed_torigoya_content_r << std::endl;
*/
    // construct envs
    char const* const command = callback_executable_r;
    char *exargv[] = {
        callback_executable_r,
        nullptr
    };

    static auto const envs_num = 1;
    std::array<std::pair<std::string, char const* const>, envs_num> envs = {
        std::make_pair( "packed_torigoya_content=", packed_torigoya_content_r )
    };

    char *exenvp[envs_num+1] = {};
    for( std::size_t i=0; i<envs.size(); ++i ) {
        char* const p = new char[envs[i].first.size() + std::strlen( envs[i].second ) + 1];
        std::copy( envs[i].first.cbegin(), envs[i].first.cend(), p );
        std::copy( envs[i].second, envs[i].second + std::strlen( envs[i].second ), p + envs[i].first.size() );
        p[envs[i].first.size() + std::strlen( envs[i].second )] = '\0'; // NULL terminate

        exenvp[i] = p;
    }
    exenvp[envs_num] = nullptr; // NULL terminate

    // Execute!
    return ::execve( command, exargv, exenvp );
}


// entry
int main( int argc, char* argv[] )
{
    // stack size: 8KBytes
    std::size_t const stack_for_child_size = 8 * 1024;
    auto const stack_for_child =
        std::make_shared<std::array<char, stack_for_child_size>>();

    if ( argc != 1 )
        return -1;

    //
    pid_t const child_pid = ::clone( fork_shell,
                                     stack_for_child->data() + stack_for_child->size(),
                                     CLONE_NEWPID | CLONE_NEWNS | CLONE_NEWNET | CLONE_NEWIPC | CLONE_NEWUTS | SIGCHLD | CLONE_UNTRACED,
                                     nullptr );
    if ( child_pid == -1 ) {
        fprintf(stderr, "Clone failed. PID namespaces ARE NOT supported\n");
        return -1;
    }
    printf( "SANDBOX: fork end - parents - PID: %d\n", getpid() );

    //
    int status;
    if ( ::waitpid( child_pid, &status, 0 ) == -1 ) {
        fprintf(stderr, "waitpid failed\n");
        return -1;
    }

    printf( "&&&&&&&&&&&&&&&& exit status code: %d\n", status );

    if ( status == 0 ) {
        return 0;

    } else {
        return -1;
    }
}
