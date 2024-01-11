local key = KEYS[1]  -- Key
local cntKey = key .. ":cnt"  -- 验证次数
local expectedCode = ARGV[1]  -- 准备存储的验证码


local cnt = tonumber(redis.call("get", cntKey))  -- 还可以验证的次数
local code = redis.call("get", key)  -- 原本存储的验证码


if cnt == nil or cnt <= 0 then
    -- 验证次数耗尽了或者系统错误了
    return -1


end

if code == expectedCode then
    redis.call("set", cntKey, 0)
    return 0
else
    -- 不相等，用户输错了
    redis.call("decr", cntKey) -- 可验证次数减1
    return -2
end



