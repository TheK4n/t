package storage

import "io"

type TasksStorage interface {
	GetNamespaces() ([]string, error)
	Count(namespace string) (uint, error)
	GetSorted(namespace string) ([]string, error)
	GetContentByIndex(namespace string, index string) ([]byte, error)
	GetContentByName(namespace string, name string) ([]byte, error)
	GetNameByIndex(namespace string, index string) (string, error)
	DeleteByIndexes(namespace string, indexes []string) error
	WriteByName(namespace string, name string, r io.Reader) error
	WriteByIndex(namespace string, index string, r io.Reader) error
	Add(namespace string, name string) error
	CountLines(namespace string, name string) (uint, error)
}
