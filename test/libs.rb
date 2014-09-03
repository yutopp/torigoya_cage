require 'yaml'
require 'colorize'
require 'torigoya_kit'

class Undefined
end

def read_setting(t)
  command_line              = if t.has_key?('command_line') then t['command_line'] else "" end
  structured_command_line   = if t.has_key?('structured_command_line') then t['structured_command_line'] else [] end
  cpu                       = if t.has_key?('cpu') then t['cpu'].to_i else raise "cpu is required" end
  memory                    = if t.has_key?('memory') then t['memory'].to_i else raise "memory is required" end

  return TorigoyaKit::ExecutionSetting.new(command_line, structured_command_line, cpu, memory)
end

def read_result(t)
  stdout        = if t.has_key?('stdout') then t['stdout'] else Undefined.new end
  stderr        = if t.has_key?('stderr') then t['stderr'] else Undefined.new end

  status        = if t.has_key?('status')       then t['status']        else Undefined.new end
  cpu           = if t.has_key?('cpu')          then t['cpu']           else Undefined.new end
  memory        = if t.has_key?('memory')       then t['memory']        else Undefined.new end
  signal        = if t.has_key?('signal')       then t['signal']        else Undefined.new end
  exit          = if t.has_key?('exit')         then t['exit']          else Undefined.new end
  command_line  = if t.has_key?('command_line') then t['command_line']  else Undefined.new end
  system_error  = if t.has_key?('system_error') then t['system_error']  else Undefined.new end

  result = TorigoyaKit::ExecutedResult.new(cpu, memory, signal, exit, command_line, status, system_error)

  return TorigoyaKit::TicketResultUnit.new(stdout, stderr, result)
end


class TicketTest
  def initialize(case_name, result, expected)
    @passed = 0
    @skipped = 0
    @failed = 0
    assert_ticket(case_name, result, expected)
  end

  private
  def assert(key, r, e)
    if e.is_a?(Undefined)
      puts "SKIPPED: #{key} is not specified".yellow
      @skipped += 1
    else
      if r != e
        puts "FAILED: #{key}: result (#{r}) but expected (#{e})".red
        @failed += 1
      else
        puts "PASSED : #{key}".green
        @passed += 1
      end
    end
  end

  def assert_result(r, e)
    assert("result/status", r.status, e.status)
    assert("result/cpu", r.used_cpu_time_sec, e.used_cpu_time_sec)
    assert("result/memoy", r.used_memory_bytes, e.used_memory_bytes)
    assert("result/signal", r.signal, e.signal)
    assert("result/exit", r.return_code, e.return_code)
    assert("result/command_line", r.command_line, e.command_line)
    assert("result/system_error", r.system_error_message, e.system_error_message)
  end

  def assert_result_unit(r, e)
    assert("out", r.out, e.out)
    assert("err", r.err, e.err)

    if e.result.nil?
      puts "SKIPPED: result is not specified".yellow
      @skipped += 1
    else
      assert_result(r.result, e.result)
    end
  end

  def assert_ticket(case_name, result, expected)
    puts ("==============" + "=" * case_name.length).blue
    puts "===>> START [".blue + case_name + "]".blue
    start_t = Time.now

    # compile
    unless expected.compile.nil?
      puts "== checking compile section".blue
      unless result.compile.nil?
        assert_result_unit(result.compile, expected.compile)
      else
        puts "FAILED: result of compile is nil".red
        @failed += 1
      end
    else
      puts "== skipped compile section".yellow
      @skipped += 1
    end

    # link
    unless expected.link.nil?
      puts "== checking link section".blue
      unless result.link.nil?
        assert_result_unit(result.link, expected.link)
      else
        puts "FAILED: result of link is nil".red
        @failed += 1
      end
    else
      puts "== skipped link section".yellow
      @skipped += 1
    end

    # run
    unless expected.run.nil?
      puts "== checking run section".blue
      unless result.run.nil?
        if result.run.size == expected.run.size
          expected.run.each do |(k, v)|
            puts "== checking run section [#{k}]".blue
            if result.run.has_key?(k)
              unless result.run[k].nil?
                assert_result_unit(result.run[k], v)
              else
                puts "FAILED: result of run [#{k}] is nil".red
                @failed += 1
              end
            else
              puts "FAILED: result has not expected index [#{k}]".red
              @failed += 1
            end
          end
        else
          puts "FAILED: numbers of run values are different(result[#{result.run.size}] != expected[#{expected.run.size}])".red
          @failed += 1
        end

      else
        puts "FAILED: result of run is nil".red
        @failed += 1
      end

    else
      puts "== skipped run section".yellow
      @skipped += 1
    end

    finish_t = Time.now
    puts "<<=== [".blue + (@failed == 0 ? "OK".green : "FAILED".red) + "]".blue + " / passed: #{@passed}".green + " / failed: #{@failed}".red + " / skipped: #{@skipped}".yellow + " -- #{finish_t-start_t} sec"
    puts ""
  end
end



def assert_ticket(case_name, result, expected)
  TicketTest.new(case_name, result, expected)
end
