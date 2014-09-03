require_relative 'run_test'

pid = nil
Signal.trap(:USR1) do
  run_test()

  Process.kill("INT", pid)
end

mode = `./host.mode.sh t`.chomp
puts "Current mode is '#{mode}'"

if mode != "debug" && mode != "release"
  puts "Error: This mode is not supported."
  puts "       Please execute 'host.mode.sh'."
  exit -1
end

system("./host.mode.sh debug")

pid = spawn("sudo bin/cage.server --mode system_test_mode --pid #{Process.pid}")
Process.waitpid(pid)

system("./host.mode.sh #{mode}")
