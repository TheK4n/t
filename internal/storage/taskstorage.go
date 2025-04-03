package storage

import "io"

type TasksStorage interface {
	GetNamespaces() ([]string, error)
	Count(namespace string) (int, error)
	GetSorted(namespace string) ([]string, error)
	GetContentByIndex(namespace string, index int) ([]byte, error)
	GetContentByName(namespace string, name string) ([]byte, error)
	GetNameByIndex(namespace string, index int) (string, error)
	DeleteByIndexes(namespace string, indexes []int) error
	WriteByName(namespace string, name string, r io.Reader) error
	WriteByIndex(namespace string, index int, r io.Reader) error
	Add(namespace string, name string) error
	CountLines(namespace string, name string) (int, error)
}
