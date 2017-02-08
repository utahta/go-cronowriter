package writer

type noopMutex struct{}

func (*noopMutex) Lock()   {}
func (*noopMutex) Unlock() {}
