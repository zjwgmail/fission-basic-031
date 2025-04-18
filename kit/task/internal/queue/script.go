package queue

import (
	"fmt"

	"github.com/go-redis/redis"

	"fission-basic/kit/task/internal/util"
)

var (
	sendScriptStr = `
local n = 0
local force = tonumber(ARGV[1]) == 1

local hash_tag = KEYS[1]:match("{([^}]+)}") or KEYS[1]
local key = "{" .. hash_tag .. "}" .. KEYS[1]

local setkey = key .. ":set"
local listkey = key .. ":list"
for i = 2, #ARGV do
	if redis.call("SADD", setkey, ARGV[i]) == 1 or force then
		if redis.call("%s", listkey, ARGV[i]) > 0 then
			n = n + 1
		end
    end
end
return n
	`
	lsendScript   = redis.NewScript(fmt.Sprintf(sendScriptStr, "LPUSH"))
	rsendScript   = redis.NewScript(fmt.Sprintf(sendScriptStr, "RPUSH"))
	receiveScript = redis.NewScript(util.ImportKitLib(`
	local vals = {}
	local interrupt = kit.interrupt(tonumber(ARGV[1]))
	local hash_tag = KEYS[1]:match("{([^}]+)}") or KEYS[1]

	for i = 1, #KEYS do
 	   	local key = "{" .. hash_tag .. "}" .. KEYS[i] .. ":list"
    	local num = kit.offset(tonumber(ARGV[i+1]))
    	local set = redis.call("LRANGE", key, 0, num - 1)
    	if #set > 0 then
     	   redis.call("LTRIM", key, num, -1)
    	end
    	kit.insert(vals, KEYS[i], set)
    	if interrupt(#set) then
        	break
    	end
	end

	return vals
	`))

	releaseScript = redis.NewScript(util.ImportKitLib(`
	local hash_tag = KEYS[1]:match("{([^}]+)}") or KEYS[1]
	local key = "{" .. hash_tag .. "}" .. KEYS[1]

	return kit.srem(key .. ":set", ARGV)
	`))
	delScript = redis.NewScript(util.ImportKitLib(`
	local hash_tag = KEYS[1]:match("{([^}]+)}") or KEYS[1]
	local key_list = "{" .. hash_tag .. "}" .. KEYS[1] .. ":list"

	for i = 1, #ARGV do
		redis.call("LREM", key_list, 0, ARGV[i])
	end

	local key_set = "{" .. hash_tag .. "}" .. KEYS[1] .. ":set"
	return kit.srem(key_set, ARGV)
	`))
	lenScript = redis.NewScript(`
	local hash_tag = KEYS[1]:match("{([^}]+)}") or KEYS[1]
	local key_list = "{" .. hash_tag .. "}" .. KEYS[1] .. ":list"

	return redis.call("LLEN", key_list) 
	`)
)
