package storage


type TasksStorage interface {
	GetSorted(namespace string) ([]string, error)
	GetContentByIndex(namespace string, index string) ([]byte, error)
	GetContentByName(namespace string, name string) ([]byte, error)
	DeleteByIndexes(namespace string, indexes []string) error
	EditByName(namespace string, name string, data []byte) error
	EditByIndex(namespace string, index string, data []byte) error
	Add(namespace string, name string) error
}