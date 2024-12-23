//go:build tsqlite

package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"strconv"
)

type SqlTasksStorage struct {
	DbPath string
}

func (ts *SqlTasksStorage) GetNamespaces() ([]string, error) {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select distinct namespace from tasks;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namespaces := []string{}

	for rows.Next() {
		namespace := ""
		err := rows.Scan(&namespace)
		if err != nil {
			fmt.Println(err)
			continue
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (ts *SqlTasksStorage) Count(namespace string) (uint, error) {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	row := db.QueryRow("select COUNT(1) from tasks where namespace = $1;", namespace)

	namespacesCount := 0
	err = row.Scan(&namespacesCount)
	if err != nil {
		return 0, err
	}

	return uint(namespacesCount), nil
}

func (ts *SqlTasksStorage) GetSorted(namespace string) ([]string, error) {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select name from tasks where namespace = $1 order by born desc;", namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []string{}

	for rows.Next() {
		task := ""
		err := rows.Scan(&task)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (ts *SqlTasksStorage) Add(namespace string, name string) error {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("insert into tasks(name, namespace, content) values($1, $2, '');", name, namespace)

	if err != nil {
		return err
	}
	return nil
}

func (ts *SqlTasksStorage) GetNameByIndex(namespace string, index string) (string, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return "", err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return "", fmt.Errorf("Wrong task index: %s", index)
	}

	return tasks[taskIndex-1], nil
}

func (ts *SqlTasksStorage) GetContentByName(namespace string, name string) ([]byte, error) {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("select content from tasks where namespace = $1 and name = $2;", namespace, name)

	taskContent := ""
	err = row.Scan(&taskContent)
	if err != nil {
		return nil, err
	}

	return []byte(taskContent), nil
}

func (ts *SqlTasksStorage) GetContentByIndex(namespace string, index string) ([]byte, error) {
	taskNameToRead, err := ts.GetNameByIndex(namespace, index)
	if err != nil {
		return nil, err
	}

	content, err := ts.GetContentByName(namespace, taskNameToRead)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (ts *SqlTasksStorage) WriteByName(namespace string, name string, r io.Reader) error {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	_, err = db.Exec("update tasks set content = $1, born = CURRENT_TIMESTAMP where name = $2 and namespace = $3;", string(b), name, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (ts *SqlTasksStorage) WriteByIndex(namespace string, index string, r io.Reader) error {
	taskNameToWrite, err := ts.GetNameByIndex(namespace, index)
	if err != nil {
		return err
	}

	return ts.WriteByName(namespace, taskNameToWrite, r)
}

func (ts *SqlTasksStorage) CountLines(namespace string, name string) (uint, error) {
	content, err := ts.GetContentByName(namespace, name)
	if err != nil {
		return 0, err
	}

	return countRune(string(content), '\n'), nil
}

func countRune(s string, r rune) uint {
	var count uint = 0
	for _, c := range s {
		if c == r {
			count++
		}
	}
	return count
}

func (ts *SqlTasksStorage) DeleteByIndexes(namespace string, indexes []string) error {
	db, err := sql.Open("sqlite3", ts.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	names := []string{}
	for _, index := range indexes {
		name, err := ts.GetNameByIndex(namespace, index)
		if err != nil {
			return err
		}
		names = append(names, name)
	}

	for _, name := range names {
		_, err = db.Exec("delete from tasks where name = $1 and namespace = $2", name, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}
