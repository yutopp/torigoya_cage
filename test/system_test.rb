require_relative 'run_test'

pid = nil
Signal.trap(:USR1) do
  run_test()

  Process.kill("KILL", pid)
end

pid = spawn("sudo bin/cage.server --mode system_test_mode --pid #{Process.pid}")
Process.waitpid(pid)
