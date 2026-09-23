package agent

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestJobQueueRetriesDeadLettersAndReplays(t *testing.T) {
	root := t.TempDir()
	queue, err := NewJobQueue(root)
	if err != nil {
		t.Fatal(err)
	}
	job, err := queue.Enqueue("mission-1", 2)
	if err != nil {
		t.Fatal(err)
	}
	claimed, ok, err := queue.Claim("worker-1", time.Now().UTC())
	if err != nil || !ok {
		t.Fatalf("claim = %+v, %v, %v", claimed, ok, err)
	}
	if _, err := queue.Nack(job.ID, errors.New("temporary")); err != nil {
		t.Fatal(err)
	}
	queue.mu.Lock()
	retry := queue.jobs[job.ID]
	retry.AvailableAt = time.Now().UTC()
	queue.jobs[job.ID] = retry
	_ = queue.persistLocked(retry)
	queue.mu.Unlock()
	claimed, ok, err = queue.Claim("worker-2", time.Now().UTC())
	if err != nil || !ok {
		t.Fatalf("second claim = %+v, %v, %v", claimed, ok, err)
	}
	dead, err := queue.Nack(job.ID, errors.New("permanent"))
	if err != nil || dead.Status != QueueDeadLetter {
		t.Fatalf("dead = %+v, err=%v", dead, err)
	}
	replayed, err := queue.Replay(job.ID)
	if err != nil || replayed.Status != QueuePending || replayed.Attempts != 0 {
		t.Fatalf("replayed = %+v, err=%v", replayed, err)
	}
	reloaded, err := NewJobQueue(root)
	if err != nil {
		t.Fatal(err)
	}
	if jobs := reloaded.List(QueuePending); len(jobs) != 1 {
		t.Fatalf("pending after reload = %+v", jobs)
	}
}

func TestJobQueueWorkerAcknowledgesJobs(t *testing.T) {
	queue, err := NewJobQueue("")
	if err != nil {
		t.Fatal(err)
	}
	job, err := queue.Enqueue("mission-2", 1)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	queue.Start(ctx, "worker", func(_ context.Context, got QueueJob) error {
		if got.ID != job.ID {
			t.Errorf("got job %+v", got)
		}
		close(done)
		return nil
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not claim job")
	}
	if status := queue.List(QueueSucceeded); len(status) != 1 {
		t.Fatalf("succeeded = %+v", status)
	}
}

func TestJobQueueRollsBackClaimWhenPersistenceFails(t *testing.T) {
	root := t.TempDir()
	queue, err := NewJobQueue(root)
	if err != nil {
		t.Fatal(err)
	}
	job, err := queue.Enqueue("mission-rollback", 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := queue.Claim("worker-rollback", time.Now().UTC()); err == nil || ok {
		t.Fatalf("claim = ok=%v err=%v, want persistence failure", ok, err)
	}
	pending := queue.List(QueuePending)
	if len(pending) != 1 || pending[0].ID != job.ID || pending[0].Attempts != 0 {
		t.Fatalf("claim mutation was not rolled back: %+v", pending)
	}
}

func TestJobQueueRejectsAckAndNackForNonRunningJobs(t *testing.T) {
	queue, err := NewJobQueue("")
	if err != nil {
		t.Fatal(err)
	}
	job, err := queue.Enqueue("mission-state", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := queue.Ack(job.ID); err == nil {
		t.Fatal("ack of pending job unexpectedly succeeded")
	}
	if _, err := queue.Nack(job.ID, errors.New("unexpected")); err == nil {
		t.Fatal("nack of pending job unexpectedly succeeded")
	}
}
