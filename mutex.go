package writer

type NoMutex struct{}

func (l *NoMutex) Lock()   {}
func (l *NoMutex) Unlock() {}
