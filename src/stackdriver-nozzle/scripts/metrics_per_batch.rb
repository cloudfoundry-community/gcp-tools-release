require "json"

count = 0
requests = 0
ARGF.each_line.each do |line|
  data = JSON.parse(line)["data"]
  data["counters"] ||= {}
  count += data["counters"].fetch("metrics.count", 0)
  requests += data["counters"].fetch("metrics.requests", 0)
  puts "#{count}/#{requests}: #{count / requests.to_f}"
end
