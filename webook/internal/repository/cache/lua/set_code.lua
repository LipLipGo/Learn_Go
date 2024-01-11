local key = KEYS[1]  -- Key
local cntKey = key .. ":cnt"  -- 验证次数
local val = ARGV[1]  -- 准备存储的验证码
-- 验证码的有效时间是10分钟
local ttl = tonumber(redis.call("ttl", key)) -- 调用ttl命令获取Key的生存时间（从当前时间开始，Key还有多少秒会过期）

if ttl == -1 then
    -- key存在，但没有过期时间
    return -2
elseif ttl == -2 or ttl < 540 then
    -- 可以发验证码
    redis.call("set", key, val)

    -- 刷新 600 秒
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)    -- 设置验证次数
    redis.call("expire", cntKey, 600)   -- 验证次数和验证码一起刷新
    return 0
else
    -- 发送太频繁
    return -1

end
