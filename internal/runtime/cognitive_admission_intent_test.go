package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	laputaevolution "github.com/ProjectViVy/laputa/laputa/evolution"

	"agent-vivy/internal/domain"
	"agent-vivy/internal/storage"
	"agent-vivy/internal/storage/sqlite"
)

// cognitiveCrashAfterAdmission fails the first cognitive state save that
// happens once a workflow run exists: the process "dies" after the workflow
// was admitted but before the attempt recorded it.
type cognitiveCrashAfterAdmission struct {
	storage.SnapshotStore
	admitted func() bool
	mu       sync.Mutex
	fired    bool
}

func (c *cognitiveCrashAfterAdmission) Put(ctx context.Context, key string, value []byte, expect int64) error {
	c.mu.Lock()
	fire := !c.fired && key == cognitiveStateKey && c.admitted()
	if fire {
		c.fired = true
	}
	c.mu.Unlock()
	if fire {
		return errors.New("fixture: process lost before the admission was recorded")
	}
	return c.SnapshotStore.Put(ctx, key, value, expect)
}

func cognitiveIntentService(t *testing.T, d *fakeCognitiveDomain) (*Service, *sqlite.Backend) {
	t.Helper()
	svc, backend := inofyExecService(t, cognitiveTestModel())
	store := &cognitiveCrashAfterAdmission{SnapshotStore: backend.Snapshot()}
	store.admitted = func() bool { return len(listWorkflowRuns(t, svc, backend)) > 0 }
	svc.deps.Cognitive = cognitiveBinding(d, &fakeSource{high: 5}, store, func() int64 { return 1000 })
	svc.StartCognitiveLoop(context.Background(), time.Hour)
	t.Cleanup(svc.StopCognitiveLoop)
	return svc, backend
}

// crashFirstAdmission runs one attempt that admits a workflow and then loses
// its state save, and waits for that workflow to reach want.
func crashFirstAdmission(t *testing.T, svc *Service, backend *sqlite.Backend, want domain.RunStatus) domain.Run {
	t.Helper()
	if _, err := svc.cognitiveAttempt(context.Background(), false); err == nil {
		t.Fatal("first attempt recorded its admission; the crash fixture did not fire")
	}
	runs := listWorkflowRuns(t, svc, backend)
	if len(runs) != 1 {
		t.Fatalf("workflow runs after the crash = %d, want 1", len(runs))
	}
	waitForRunStatus(t, backend, runs[0].ID, want)
	return runs[0]
}

func TestCognitiveCrashAfterAdmissionAdoptsCompletedWindow(t *testing.T) {
	ctx := context.Background()
	svc, backend := cognitiveIntentService(t, &fakeCognitiveDomain{batch: laputaevolution.EvidenceBatch{ActivityRevision: 3}})
	crashFirstAdmission(t, svc, backend, domain.RunCompleted)

	elig, err := svc.cognitiveAttempt(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if elig.Run || elig.Reason != laputaevolution.ReasonNoNewInput {
		t.Fatalf("attempt after the crash = %+v, want no_new_input", elig)
	}
	st, _, err := svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.Watermark != 5 || st.Attempt != 0 || st.ActiveRunID != "" || st.IntentOpKey != "" {
		t.Fatalf("state after adopting the completed window = %+v", st)
	}
	if n := len(listWorkflowRuns(t, svc, backend)); n != 1 {
		t.Fatalf("completed window was admitted again: %d workflow runs", n)
	}
}

// The restart shape: recovery has already closed the orphaned supervisor, so
// the next attempt parents new work elsewhere. The intent still names the
// exact parent and key of the admitted run.
func TestCognitiveCrashAfterAdmissionSurvivesSupervisorReplacement(t *testing.T) {
	ctx := context.Background()
	svc, backend := cognitiveIntentService(t, &fakeCognitiveDomain{batch: laputaevolution.EvidenceBatch{ActivityRevision: 3}})
	wf := crashFirstAdmission(t, svc, backend, domain.RunCompleted)
	if err := backend.SetRunStatus(ctx, wf.ParentID, domain.RunFailed); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.cognitiveAttempt(ctx, false); err != nil {
		t.Fatal(err)
	}
	st, _, err := svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.Watermark != 5 || st.Attempt != 0 {
		t.Fatalf("state after supervisor replacement = %+v", st)
	}
	if n := len(listWorkflowRuns(t, svc, backend)); n != 1 {
		t.Fatalf("completed window was admitted again: %d workflow runs", n)
	}
}

// A transiently failed run keeps the existing retry contract: one attempt is
// counted and the window is retried under the next attempt key.
func TestCognitiveCrashAfterAdmissionRetriesFailedWindowOnce(t *testing.T) {
	ctx := context.Background()
	svc, backend := cognitiveIntentService(t, &fakeCognitiveDomain{
		batch:      laputaevolution.EvidenceBatch{ActivityRevision: 3},
		collectErr: errors.New("fixture: collect exploded"),
	})
	first := crashFirstAdmission(t, svc, backend, domain.RunFailed)

	elig, err := svc.cognitiveAttempt(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if !elig.Run {
		t.Fatalf("failed window was not retried: %+v", elig)
	}
	st, _, err := svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.Attempt != 1 || st.ActiveRunID == "" || st.ActiveRunID == string(first.ID) {
		t.Fatalf("state after retry = %+v", st)
	}
	if _, err := backend.GetWorkflowRevisionByOperation(ctx, domain.RunID(st.SupervisorRunID), "cognitive:activity:0-5:a1"); err != nil {
		t.Fatalf("retry was not admitted under attempt a1: %v", err)
	}
}

// An intent whose admission never committed means nothing ran: it is
// dropped and the window is admitted normally.
func TestCognitiveIntentWithoutAdmissionIsDropped(t *testing.T) {
	ctx := context.Background()
	svc, backend := inofyExecService(t, cognitiveTestModel())
	svc.deps.Cognitive = cognitiveBinding(&fakeCognitiveDomain{batch: laputaevolution.EvidenceBatch{ActivityRevision: 3}},
		&fakeSource{high: 5}, backend.Snapshot(), func() int64 { return 1000 })
	svc.StartCognitiveLoop(ctx, time.Hour)
	t.Cleanup(svc.StopCognitiveLoop)
	st, version, err := svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	st.IntentOpKey, st.IntentParentRunID, st.IntentThrough = "cognitive:activity:0-5:a0", "run_never_created", 5
	if err := svc.saveCognitiveState(ctx, st, version); err != nil {
		t.Fatal(err)
	}

	elig, err := svc.cognitiveAttempt(ctx, false)
	if err != nil || !elig.Run {
		t.Fatalf("attempt with an uncommitted intent = %+v, %v", elig, err)
	}
	if n := len(listWorkflowRuns(t, svc, backend)); n != 1 {
		t.Fatalf("workflow runs = %d, want 1", n)
	}
	st, _, err = svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.IntentOpKey != "" || st.ActiveRunID == "" {
		t.Fatalf("state after admission = %+v", st)
	}
}

// A supervisor created just before a crash keeps its id, but the sequence
// that named it was never saved. The next supervisor must not reuse the id.
func TestCognitiveSupervisorSkipsUnsavedSequence(t *testing.T) {
	ctx := context.Background()
	svc, backend := inofyExecService(t, cognitiveTestModel())
	svc.deps.Cognitive = cognitiveBinding(&fakeCognitiveDomain{batch: laputaevolution.EvidenceBatch{ActivityRevision: 3}},
		&fakeSource{high: 5}, backend.Snapshot(), func() int64 { return 1000 })
	svc.StartCognitiveLoop(ctx, time.Hour)
	t.Cleanup(svc.StopCognitiveLoop)
	if err := backend.CreateSession(ctx, domain.Session{ID: cognitiveSupervisorSessionID, Title: "cognitive supervisor", CreatedAt: 1}); err != nil {
		t.Fatal(err)
	}
	orphan := domain.RunID("run_cognitive_supervisor_1")
	if err := backend.CreateRun(ctx, domain.Run{ID: orphan, SessionID: cognitiveSupervisorSessionID, Status: domain.RunFailed,
		Kind: domain.RunKindPrimary, CreatedAt: 1, RootID: orphan}); err != nil {
		t.Fatal(err)
	}

	if elig, err := svc.cognitiveAttempt(ctx, false); err != nil || !elig.Run {
		t.Fatalf("attempt after an orphaned supervisor = %+v, %v", elig, err)
	}
	st, _, err := svc.loadCognitiveState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.SupervisorRunID == string(orphan) || st.SupervisorRunID == "" {
		t.Fatalf("supervisor = %q, want a fresh id", st.SupervisorRunID)
	}
}
