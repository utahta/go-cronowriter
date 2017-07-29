package cronowriter

import "testing"

func TestNopMutex(t *testing.T) {
	m := &nopMutex{}
	m.Lock()
	m.Unlock()
}
