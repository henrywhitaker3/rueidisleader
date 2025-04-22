local topic = KEYS[1]
local instance = ARGV[1]
local lifetime = ARGV[2]

if redis.call("GET", topic) == instance then
	redis.call("PEXPIRE", topic, lifetime)
	return true
end

return false
