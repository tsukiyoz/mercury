local fieldKeys = KEYS
local cntKey = ARGV[1]
local delta = tonumber(ARGV[2])

local result = {}
for i, key in ipairs(fieldKeys) do
    local exists = redis.call("EXISTS", key)
    if exists == 1 then
        redis.call("HINCRBY", key, cntKey, delta)
        table.insert(result, 1)
    else
        table.insert(result, 0)
    end
end

return result