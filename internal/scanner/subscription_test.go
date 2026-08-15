package scanner

import "testing"

func TestUnsubscribeIsIdempotentAfterCompletion(t *testing.T) {
	manager := &Manager{jobs: map[string]*job{}}
	running := &job{subscribers: map[chan Progress]struct{}{}}
	manager.jobs["scan"] = running
	_, unsubscribe, err := manager.Subscribe("scan")
	if err != nil {
		t.Fatal(err)
	}
	unsubscribe()
	unsubscribe()
}
