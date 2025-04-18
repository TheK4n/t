<h1 align="center">T: simple task tracker</h1>

<p align="center">
  <a href="https://github.com/TheK4n">
    <img src="https://img.shields.io/github/followers/TheK4n?label=Follow&style=social">
  </a>
  <a href="https://github.com/TheK4n/t">
    <img src="https://img.shields.io/github/stars/TheK4n/t?style=social">
  </a>
</p>

* [Features](#features)
* [Installation](#installation)
* [Usage](#usage)

---

Simple task tracker


## Installation

### Dependencies

Build dependencies:
* golang


### Compile from source:
```sh
git clone https://github.com/thek4n/t.git
cd t
go build ./cmd/t
```

### Install by golang (recommended):
```sh
go install github.com/thek4n/t/cmd/t@v1.3.4
```

### Download binary
```sh
wget https://github.com/TheK4n/t/releases/download/v1.3.4/t_v1.3.4_linux_amd64.tar.gz
tar xzf t_v1.3.4_linux_amd64.tar.gz
```


## Usage
```sh
t --help
t a Buy bread  # Add task
t      # Show tasks
t e 1  # Edit task content
t 1    # Show task content with index 1
```


### Install with sqlite support
Sqlite3 db as storage instead of files

```sh
go install --tags=tsqlite github.com/thek4n/t/cmd/t@v1.3.4
t -v  # v1.3.4-sqlite
```
