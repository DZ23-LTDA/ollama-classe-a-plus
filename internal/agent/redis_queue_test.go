package agent

import (
	"context"
	"strings"
	"testing"
)

func TestOpenRedisQueueRejectsRedissWithoutTLSAdapter(t *testing.T) {
	_, err := OpenRedisQueue(context.Background(), "rediss://127.0.0.1:1/0", "test")
	if err == nil || !strings.Contains(err.Error(), "TLS") {
		t.Fatalf("rediss error = %v", err)
	}
}

func TestRedisClaimAndReclaimUseAtomicLeaseScripts(t *testing.T) {
	for name, script := range map[string]string{"claim": redisClaimScript, "reclaim": redisReclaimScript} {
		if !strings.Contains(script, "redis.call") || !strings.Contains(script, "job.status") {
			t.Fatalf("%s script is not a Redis state transition", name)
		}
	}
	if redisLeaseDuration <= 0 {
		t.Fatal("redis lease duration must be positive")
	}
}
