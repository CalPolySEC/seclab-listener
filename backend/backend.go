package backend

type Backend interface {
	Open() error
	Close() error
}

type fileBackend struct {
	linkPath   string
	openPath   string
	closedPath string
}

func New(linkPath, openPath, closedPath string) Backend {
	return &fileBackend{
		linkPath:   linkPath,
		openPath:   openPath,
		closedPath: closedPath,
	}
}

func (b *fileBackend) Open() error {
	return nil
}

func (b *fileBackend) Close() error {
	return nil
}
