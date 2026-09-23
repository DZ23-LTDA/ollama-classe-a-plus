package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RedisQueue struct {
	address, password string
	database          int
	prefix            string
	timeout           time.Duration
	errors            chan error
}

const redisLeaseDuration = 15 * time.Minute

const redisClaimScript = `
local id = redis.call('RPOP', KEYS[1])
if not id then return nil end
local raw = redis.call('GET', KEYS[4] .. id)
if not raw then return nil end
local job = cjson.decode(raw)
if job.status ~= 'pending' then return nil end
job.status = 'running'
job.attempts = (job.attempts or 0) + 1
job.worker_id = ARGV[1]
job.locked_at = ARGV[2]
job.updated_at = ARGV[2]
redis.call('SET', KEYS[4] .. id, cjson.encode(job), 'EX', '604800')
redis.call('ZADD', KEYS[3], ARGV[3] + ARGV[4], id)
return cjson.encode(job)
`

const redisReclaimScript = `
local ids = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
local reclaimed = 0
for _, id in ipairs(ids) do
  redis.call('ZREM', KEYS[1], id)
  local raw = redis.call('GET', KEYS[2] .. id)
  if raw then
    local job = cjson.decode(raw)
    if job.status == 'running' then
      job.status = 'pending'
      job.worker_id = ''
      job.locked_at = cjson.null
      job.available_at = ARGV[2]
      job.updated_at = ARGV[2]
      redis.call('SET', KEYS[2] .. id, cjson.encode(job), 'EX', '604800')
      redis.call('LPUSH', ARGV[3], id)
      reclaimed = reclaimed + 1
    end
  end
end
return reclaimed
`

const redisEnqueueScript = `
local existingID = redis.call('GET', KEYS[1])
if existingID then
  local existingRaw = redis.call('GET', KEYS[2] .. existingID)
  if existingRaw then
    local existing = cjson.decode(existingRaw)
    if existing.mission_id == ARGV[2] and (existing.status == 'pending' or existing.status == 'running') then
      return existingRaw
    end
  end
  redis.call('DEL', KEYS[1])
end
redis.call('SET', KEYS[2] .. cjson.decode(ARGV[1]).id, ARGV[1], 'EX', '604800')
redis.call('SET', KEYS[1], cjson.decode(ARGV[1]).id, 'EX', '604800')
redis.call('LPUSH', KEYS[3], cjson.decode(ARGV[1]).id)
return ARGV[1]
`

func OpenRedisQueue(ctx context.Context, rawURL, prefix string) (*RedisQueue, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return nil, errors.New("invalid Redis URL")
	}
	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return nil, errors.New("Redis URL must use redis:// or rediss://")
	}
	if u.Scheme == "rediss" {
		return nil, errors.New("rediss:// requires a TLS-configured Redis adapter; plaintext fallback is disabled")
	}
	database := 0
	if value := strings.TrimPrefix(u.Path, "/"); value != "" {
		database, err = strconv.Atoi(value)
		if err != nil || database < 0 {
			return nil, errors.New("invalid Redis database")
		}
	}
	password, _ := u.User.Password()
	queue := &RedisQueue{address: u.Host, password: password, database: database, prefix: strings.TrimSuffix(prefix, ":"), timeout: 5 * time.Second, errors: make(chan error, 32)}
	if queue.prefix == "" {
		queue.prefix = "ollama:agent"
	}
	if err := queue.ping(ctx); err != nil {
		return nil, err
	}
	return queue, nil
}

func (q *RedisQueue) key(name string) string  { return q.prefix + ":" + name }
func (q *RedisQueue) jobKey(id string) string { return q.key("job:" + id) }
func (q *RedisQueue) missionKey(id string) string {
	return q.key("mission:" + id)
}
func (q *RedisQueue) pendingKey() string { return q.key("pending") }
func (q *RedisQueue) delayedKey() string { return q.key("delayed") }
func (q *RedisQueue) deadKey() string    { return q.key("dead") }
func (q *RedisQueue) leaseKey() string   { return q.key("leases") }

func (q *RedisQueue) Enqueue(missionID string, maxAttempts int) (QueueJob, error) {
	if strings.TrimSpace(missionID) == "" {
		return QueueJob{}, errors.New("mission id is required")
	}
	if maxAttempts <= 0 || maxAttempts > 20 {
		maxAttempts = 3
	}
	now := time.Now().UTC()
	job := QueueJob{ID: "job_" + uuid.NewString(), MissionID: missionID, Status: QueuePending, MaxAttempts: maxAttempts, AvailableAt: now, CreatedAt: now, UpdatedAt: now}
	data, err := json.Marshal(job)
	if err != nil {
		return QueueJob{}, err
	}
	value, err := q.do(context.Background(), "EVAL", redisEnqueueScript, "3", q.missionKey(missionID), q.key("job:"), q.pendingKey(), string(data), missionID)
	if err != nil {
		return QueueJob{}, err
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return QueueJob{}, errors.New("Redis enqueue returned an invalid job")
	}
	if err := json.Unmarshal([]byte(text), &job); err != nil {
		return QueueJob{}, err
	}
	return job, nil
}

func (q *RedisQueue) Claim(workerID string, now time.Time) (QueueJob, bool, error) {
	if strings.TrimSpace(workerID) == "" {
		return QueueJob{}, false, errors.New("worker id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := q.moveDue(context.Background(), now); err != nil {
		return QueueJob{}, false, err
	}
	if err := q.reclaimExpired(context.Background(), now); err != nil {
		return QueueJob{}, false, err
	}
	value, err := q.do(context.Background(), "EVAL", redisClaimScript, "4", q.pendingKey(), q.delayedKey(), q.leaseKey(), q.key("job:"), workerID, now.Format(time.RFC3339Nano), strconv.FormatInt(now.UnixMilli(), 10), strconv.FormatInt(redisLeaseDuration.Milliseconds(), 10))
	if err != nil {
		return QueueJob{}, false, err
	}
	if value == nil {
		return QueueJob{}, false, nil
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return QueueJob{}, false, errors.New("Redis claim returned an invalid job")
	}
	var job QueueJob
	if err := json.Unmarshal([]byte(text), &job); err != nil {
		return QueueJob{}, false, err
	}
	return job, true, nil
}

func (q *RedisQueue) Ack(jobID string) error {
	job, err := q.get(jobID)
	if err != nil {
		return err
	}
	if job.Status != QueueRunning {
		return errors.New("job is not running")
	}
	job.Status = QueueSucceeded
	job.WorkerID = ""
	job.LockedAt = nil
	job.UpdatedAt = time.Now().UTC()
	return q.put(job)
}

func (q *RedisQueue) Nack(jobID string, runErr error) (QueueJob, error) {
	job, err := q.get(jobID)
	if err != nil {
		return QueueJob{}, err
	}
	if job.Status != QueueRunning {
		return QueueJob{}, errors.New("job is not running")
	}
	previous := job
	if runErr != nil {
		job.LastError = limitError(runErr.Error(), 2000)
	}
	job.WorkerID = ""
	job.LockedAt = nil
	job.UpdatedAt = time.Now().UTC()
	if job.Attempts >= job.MaxAttempts {
		job.Status = QueueDeadLetter
		if err := q.put(job); err != nil {
			return QueueJob{}, err
		}
		if _, err := q.do(context.Background(), "LPUSH", q.deadKey(), job.ID); err != nil {
			if restoreErr := q.put(previous); restoreErr != nil {
				return QueueJob{}, fmt.Errorf("dead-letter indexing failed: %v; state restore failed: %w", err, restoreErr)
			}
			return QueueJob{}, fmt.Errorf("dead-letter indexing failed: %w", err)
		}
		return job, nil
	}
	job.Status = QueuePending
	backoff := time.Duration(1<<(job.Attempts-1)) * time.Second
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	job.AvailableAt = time.Now().UTC().Add(backoff)
	if err := q.put(job); err != nil {
		return QueueJob{}, err
	}
	if _, err = q.do(context.Background(), "ZADD", q.delayedKey(), strconv.FormatInt(job.AvailableAt.UnixMilli(), 10), job.ID); err != nil {
		if restoreErr := q.put(previous); restoreErr != nil {
			return QueueJob{}, fmt.Errorf("queue retry scheduling failed: %v; state restore failed: %w", err, restoreErr)
		}
		return QueueJob{}, err
	}
	return job, nil
}

func (q *RedisQueue) Replay(jobID string) (QueueJob, error) {
	job, err := q.get(jobID)
	if err != nil {
		return QueueJob{}, err
	}
	if job.Status != QueueDeadLetter && job.Status != QueueFailed {
		return QueueJob{}, fmt.Errorf("job %s is not replayable", jobID)
	}
	previous := job
	job.Status = QueuePending
	job.Attempts = 0
	job.LastError = ""
	job.WorkerID = ""
	job.LockedAt = nil
	job.AvailableAt = time.Now().UTC()
	job.UpdatedAt = time.Now().UTC()
	if err := q.put(job); err != nil {
		return QueueJob{}, err
	}
	_, err = q.do(context.Background(), "LPUSH", q.pendingKey(), job.ID)
	if err != nil {
		if restoreErr := q.put(previous); restoreErr != nil {
			return QueueJob{}, fmt.Errorf("queue replay enqueue failed: %v; state restore failed: %w", err, restoreErr)
		}
	}
	return job, err
}

func (q *RedisQueue) List(status QueueStatus) []QueueJob {
	value, err := q.do(context.Background(), "KEYS", q.key("job:*"))
	if err != nil {
		return nil
	}
	keys, ok := value.([]any)
	if !ok {
		return nil
	}
	jobs := []QueueJob{}
	for _, raw := range keys {
		id, ok := raw.(string)
		if !ok {
			continue
		}
		job, err := q.get(strings.TrimPrefix(id, q.key("job:")))
		if err == nil && (status == "" || job.Status == status) {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

func (q *RedisQueue) Start(ctx context.Context, workerID string, handler func(context.Context, QueueJob) error) {
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			job, ok, err := q.Claim(workerID, time.Now().UTC())
			if err != nil {
				q.reportError(fmt.Errorf("redis queue claim: %w", err))
			}
			if err == nil && ok {
				if runErr := handler(ctx, job); runErr != nil {
					if _, nackErr := q.Nack(job.ID, runErr); nackErr != nil {
						q.reportError(fmt.Errorf("redis queue nack %s: %w", job.ID, nackErr))
					}
				} else {
					if ackErr := q.Ack(job.ID); ackErr != nil {
						q.reportError(fmt.Errorf("redis queue ack %s: %w", job.ID, ackErr))
					}
				}
				continue
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}()
}

func (q *RedisQueue) Errors() <-chan error {
	if q == nil {
		return nil
	}
	return q.errors
}

func (q *RedisQueue) reportError(err error) {
	if q == nil || q.errors == nil || err == nil {
		return
	}
	select {
	case q.errors <- err:
	default:
	}
}

func (q *RedisQueue) ping(ctx context.Context) error { _, err := q.do(ctx, "PING"); return err }
func (q *RedisQueue) get(id string) (QueueJob, error) {
	value, err := q.do(context.Background(), "GET", q.jobKey(id))
	if err != nil {
		return QueueJob{}, err
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return QueueJob{}, osErrNotExist{}
	}
	var job QueueJob
	if err := json.Unmarshal([]byte(text), &job); err != nil {
		return QueueJob{}, err
	}
	return job, nil
}

func (q *RedisQueue) put(job QueueJob) error {
	data, _ := json.Marshal(job)
	_, err := q.do(context.Background(), "SET", q.jobKey(job.ID), string(data), "EX", "604800")
	return err
}

func (q *RedisQueue) moveDue(ctx context.Context, now time.Time) error {
	value, err := q.do(ctx, "ZRANGEBYSCORE", q.delayedKey(), "-inf", strconv.FormatInt(now.UnixMilli(), 10))
	if err != nil {
		return err
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	for _, raw := range items {
		id, ok := raw.(string)
		if !ok {
			continue
		}
		if _, err := q.do(ctx, "ZREM", q.delayedKey(), id); err != nil {
			return err
		}
		if _, err := q.do(ctx, "LPUSH", q.pendingKey(), id); err != nil {
			_, _ = q.do(ctx, "ZADD", q.delayedKey(), strconv.FormatInt(now.UnixMilli(), 10), id)
			return err
		}
	}
	return nil
}

func (q *RedisQueue) reclaimExpired(ctx context.Context, now time.Time) error {
	_, err := q.do(ctx, "EVAL", redisReclaimScript, "2", q.leaseKey(), q.key("job:"), strconv.FormatInt(now.UnixMilli(), 10), now.Format(time.RFC3339Nano), q.pendingKey())
	return err
}

func (q *RedisQueue) do(ctx context.Context, args ...string) (any, error) {
	timeout := q.timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", q.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if deadline, ok := dialCtx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	reader := bufio.NewReader(conn)
	if q.password != "" {
		if _, err := redisCommand(conn, reader, "AUTH", q.password); err != nil {
			return nil, err
		}
	}
	if q.database > 0 {
		if _, err := redisCommand(conn, reader, "SELECT", strconv.Itoa(q.database)); err != nil {
			return nil, err
		}
	}
	return redisCommand(conn, reader, args...)
}

func redisCommand(w io.Writer, r *bufio.Reader, args ...string) (any, error) {
	var builder strings.Builder
	builder.WriteString("*" + strconv.Itoa(len(args)) + "\r\n")
	for _, arg := range args {
		builder.WriteString("$" + strconv.Itoa(len(arg)) + "\r\n" + arg + "\r\n")
	}
	if _, err := io.WriteString(w, builder.String()); err != nil {
		return nil, err
	}
	return readRedis(r)
}

func readRedis(r *bufio.Reader) (any, error) {
	kind, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch kind {
	case '+':
		line, err := r.ReadString('\n')
		return strings.TrimSpace(line), err
	case '-':
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return nil, errors.New(strings.TrimSpace(line))
	case ':':
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return strconv.ParseInt(strings.TrimSpace(line), 10, 64)
	case '$':
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		if size < 0 {
			return nil, nil
		}
		data := make([]byte, size+2)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		return string(data[:size]), nil
	case '*':
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		if count < 0 {
			return nil, nil
		}
		items := make([]any, count)
		for index := range items {
			items[index], err = readRedis(r)
			if err != nil {
				return nil, err
			}
		}
		return items, nil
	}
	return nil, fmt.Errorf("unsupported Redis response %q", kind)
}

type osErrNotExist struct{}

func (osErrNotExist) Error() string { return "redis job not found" }
