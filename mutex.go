package cronowriter

type nopMutex struct{}

func (*nopMutex) Lock()   {}
func (*nopMutex) Unlock() {}
