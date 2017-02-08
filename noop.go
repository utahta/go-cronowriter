package writer

type noopMutex struct{}
type noopWriter struct{}

func (*noopMutex) Lock()   {}
func (*noopMutex) Unlock() {}

func (*noopWriter) Write([]byte) (int, error) {
	return 0, nil // no-op
}
