sqlite3 test.sqlite3 'CREATE TABLE namespaces(name varchar(30) primary key);'
sqlite3 test.sqlite3 'CREATE TABLE tasks(name varchar(90) primary key, namespace varchar(30) not null references namespaces(name), born timestamp default CURRENT_TIMESTAMP not null, content text);'
