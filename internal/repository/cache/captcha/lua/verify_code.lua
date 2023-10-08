local key = KEYS[1]
local expectedCaptcha = ARGV[1]
local captcha = redis.call("get", key)
local cntKey = key .. ":cnt"
local cnt = tonumber(redis.call("get", cntKey))
if cnt <= 0 then
    -- always wrong or has been used
    return -1
end

if expectedCaptcha == captcha then
    -- correct
    redis.call("set", cntKey, -1)
    return 0
else
    -- wrong captcha
    redis.call("decr", cntKey)
    return -2
end