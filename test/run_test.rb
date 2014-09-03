require 'yaml'
require 'colorize'
require 'torigoya_kit'
require 'optparse'
require_relative 'libs'

def run_test()
  #
  opt = OptionParser.new
  cases = nil
  opt.on('-c CASES', Array) {|v| cases = v }
  opt.parse!(ARGV)


  #
  c = TorigoyaKit::Client.new("localhost", 49800)
  c.update_packages()
  c.reload_proc_table()


  #
  testcases_path = File.join(File.expand_path(File.dirname(__FILE__)), "../torigoya_proc_profiles/_testcases")
  test_paths = Dir.glob(File.join(testcases_path, "*"))

  puts "Torigoya system test: run #{test_paths.length} tests"

  results = []
  failed_num = 0

  test_paths.each do |dir_name|
    Dir.chdir(dir_name) do
      unless cases.nil?
        next unless cases.any? {|c| /#{c}/ =~ dir_name}
      end

      Dir.glob(File.join("testcase*.yml")) do |unit_path|
        begin
          puts unit_path.red

          testcase = YAML.load_file(unit_path)

          # source
          sources = testcase['source_files'].map do |source_file|
            # TODO: support multiple sources...
            TorigoyaKit::SourceData.new(nil, File.read(source_file))
          end

          # compile unit
          compile_setting = nil
          compile_expect = nil
          unless testcase['compile'].nil? then
            t = testcase['compile']['try']
            compile_setting = unless t.nil? then
                                read_setting(t)
                              else
                                TorigoyaKit::ExecutionSetting.new("", [], 10, 1024*1024*1024)
                              end

            e = testcase['compile']['expect']
            unless e.nil?
              compile_expect = read_result(e)
            else
              puts "expect section is not found[compile]".yellow
            end
          end

          link_setting = nil
          link_expect = nil
          unless testcase['link'].nil? then
            t = testcase['link']['try']
            link_setting = unless t.nil? then
                             read_setting(t)
                           else
                             TorigoyaKit::ExecutionSetting.new("", [], 10, 1024*1024*1024)
                           end

            e = testcase['link']['expect']
            unless e.nil?
              link_expect = read_result(e)
            else
              puts "expect section is not found[link]".yellow
            end
          end

          build_inst = nil
          unless compile_setting.nil? && link_setting.nil?
            build_inst = TorigoyaKit::BuildInstruction.new(compile_setting, link_setting)
          end

          run_inst = nil
          run_expect = {}
          unless testcase['run'].nil? then
            inputs = []
            testcase['run'].each_with_index do |run_unit, index|
              t = run_unit['try']

              setting = unless t.nil? then
                          read_setting(t)
                        else
                          TorigoyaClient::ExecutionSetting.new("", [], 10, 512 * 1024 * 1024)
                        end

              inputs << TorigoyaKit::Input.new(t['stdin'], setting)

              e = run_unit['expect']
              unless e.nil?
                run_expect[index] = read_result(e)
              else
                puts "expect section is not found[run/#{index}]".yellow
              end
            end
            run_inst = TorigoyaKit::RunInstruction.new(inputs)
          end

          # ticket!
          ticket = TorigoyaKit::Ticket.new(unit_path,
                                           testcase['id'].to_i,
                                           testcase['version'],
                                           sources,
                                           build_inst,
                                           run_inst
                                           )
          # execute ticket
          ticket = c.exec_ticket(ticket)
          expected_ticket = TorigoyaKit::TicketResult.new(compile_expect, link_expect, run_expect)

          result, s = assert_ticket("#{File.basename(dir_name)} - #{unit_path}", ticket, expected_ticket)
          results << result
          failed_num += 1 if s == false

        rescue => e
          p e
          #=begin
          e.backtrace.each do |b|
            puts "== #{b}"
          end
        end
      end
    end
  end # test_paths.each

  puts "== ALL TEST FINISHED =="
  puts results
  puts ""
  if failed_num == 0
    puts "====> ALL GREEN!!".green
  else
    puts "====> FAILED...".red
  end
  puts "======================="

end # def run_test
