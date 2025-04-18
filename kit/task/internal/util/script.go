package util

var (
	KitLib = `
local kit = {}

kit.time = function()
	local ts = redis.call("TIME")
	if #ts == 2 then
		return tonumber(ts[1]) * 1e6 + tonumber(ts[2])
	end
	return 0
end

kit.interrupt = function(timeout)
	local now = (timeout > 0) and kit.time() or 0
	local total = 0
	return function(num)
		total = total + num
		if total > 1000 then
			return true
		end
		if timeout > 0 and kit.time() - now > timeout then
			return true
		end
		return false
	end
end

kit.offset = function(num)
	return (num <= 1000) and num or 1000
end

kit.insert = function(vals, key, set)
	table.insert(vals, key)
	table.insert(vals, tostring(#set))
	for _, v in pairs(set) do
		table.insert(vals, v)
	end
end

kit.split = function(list)
	local t = {}
	local sub = {}
	for i = 1, #list do
		table.insert(sub, list[i])	
		if #sub > 100 then
			table.insert(t, sub)
			sub = {}
		end
	end
	if #sub > 0 then
		table.insert(t, sub)
	end
	return t
end

kit.srem = function(key, members)
	local n = 0
	local t = kit.split(members)
	for i = 1, #t do
		n = n + redis.call("SREM", key, unpack(t[i]))
	end
	return n
end
	`
)
