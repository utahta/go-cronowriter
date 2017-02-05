package writer

type NoMutex struct{}

func (*NoMutex) Lock()   {}
func (*NoMutex) Unlock() {}
