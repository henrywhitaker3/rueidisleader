local topic = KEYS[1]
local instance = ARGV[1]

if redis.call("GET", topic) == instance then
	redis.call("DEL", topic)
	return true
end

return false
