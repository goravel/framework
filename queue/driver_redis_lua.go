package queue

import "github.com/redis/go-redis/v9"

// RedisLua Redis Lua Scripts
// source: https://github.com/laravel/framework/blob/10.x/src/Illuminate/Queue/LuaScripts.php
type RedisLua struct{}

// Size Get the Lua script for computing the size of queue.
// KEYS[1] - The name of the primary queue
// KEYS[2] - The name of the "delayed" queue
// KEYS[3] - The name of the "reserved" queue
func (r *RedisLua) Size() *redis.Script {
	return redis.NewScript(`
return redis.call('llen', KEYS[1]) + redis.call('zcard', KEYS[2]) + redis.call('zcard', KEYS[3])
`)
}

// Push Get the Lua script for pushing jobs onto the queue.
// KEYS[1] - The queue to push the job onto, for example: queues:foo
// KEYS[2] - The notification list for the queue we are pushing jobs onto, for example: queues:foo:notify
// ARGV[1] - The job payload
func (r *RedisLua) Push() *redis.Script {
	return redis.NewScript(`
-- Push the job onto the queue...
redis.call('rpush', KEYS[1], ARGV[1])
-- Push a notification onto the "notify" queue...
redis.call('rpush', KEYS[2], 1)
`)
}

// Pop Get the Lua script for popping the next job off of the queue.
// KEYS[1] - The queue to pop jobs from, for example: queues:foo
// KEYS[2] - The queue to place reserved jobs on, for example: queues:foo:reserved
// KEYS[3] - The notify queue
// ARGV[1] - The time at which the reserved job will expire
func (r *RedisLua) Pop() *redis.Script {
	return redis.NewScript(`
-- Pop the first job off of the queue...
local job = redis.call('lpop', KEYS[1])
local reserved = false

if(job ~= false) then
    -- Increment the attempt count and place job on the reserved queue...
    reserved = cjson.decode(job)
    reserved['attempts'] = reserved['attempts'] + 1
    reserved = cjson.encode(reserved)
    redis.call('zadd', KEYS[2], ARGV[1], reserved)
    redis.call('lpop', KEYS[3])
end

return {job, reserved}
`)
}

// Release Get the Lua script for releasing reserved jobs.
// KEYS[1] - The "delayed" queue we release jobs onto, for example: queues:foo:delayed
// KEYS[2] - The queue the jobs are currently on, for example: queues:foo:reserved
// ARGV[1] - The raw payload of the job to add to the "delayed" queue
// ARGV[2] - The UNIX timestamp at which the job should become available
func (r *RedisLua) Release() *redis.Script {
	return redis.NewScript(`
-- Remove the job from the current queue...
redis.call('zrem', KEYS[2], ARGV[1])

-- Add the job onto the "delayed" queue...
redis.call('zadd', KEYS[1], ARGV[2], ARGV[1])

return true
`)
}

// MigrateExpiredJobs Get the Lua script to migrate expired jobs back onto the queue.
// KEYS[1] - The queue we are removing jobs from, for example: queues:foo:reserved
// KEYS[2] - The queue we are moving jobs to, for example: queues:foo
// KEYS[3] - The notification list for the queue we are moving jobs to, for example queues:foo:notify
// ARGV[1] - The current UNIX timestamp
func (r *RedisLua) MigrateExpiredJobs() *redis.Script {
	return redis.NewScript(`
-- Get all of the jobs with an expired "score"...
local val = redis.call('zrangebyscore', KEYS[1], '-inf', ARGV[1], 'limit', 0, ARGV[2])

-- If we have values in the array, we will remove them from the first queue
-- and add them onto the destination queue in chunks of 100, which moves
-- all of the appropriate jobs onto the destination queue very safely.
if(next(val) ~= nil) then
    redis.call('zremrangebyrank', KEYS[1], 0, #val - 1)

    for i = 1, #val, 100 do
        redis.call('rpush', KEYS[2], unpack(val, i, math.min(i+99, #val)))
        -- Push a notification for every job that was migrated...
        for j = i, math.min(i+99, #val) do
            redis.call('rpush', KEYS[3], 1)
        end
    end
end

return val
`)
}

// Clear Get the Lua script for removing all jobs from the queue.
// KEYS[1] - The name of the primary queue
// KEYS[2] - The name of the "delayed" queue
// KEYS[3] - The name of the "reserved" queue
// KEYS[4] - The name of the "notify" queue
func (r *RedisLua) Clear() *redis.Script {
	return redis.NewScript(`
local size = redis.call('llen', KEYS[1]) + redis.call('zcard', KEYS[2]) + redis.call('zcard', KEYS[3])
redis.call('del', KEYS[1], KEYS[2], KEYS[3], KEYS[4])
return size
`)
}
