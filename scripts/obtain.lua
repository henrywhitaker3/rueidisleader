local topic = KEYS[1]
local instance = ARGV[1]
local lifetime = ARGV[2]

if redis.call("SET", topic, instance, "PX", lifetime, "NX") == false then
	return false
end

return 1
