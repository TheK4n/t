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

Dependencies:
* golang


### Compile from source:
```sh
git clone https://github.com/thek4n/t.git
cd t
go build ./cmd/t
```

### Install by golang (recommended):
```sh
go install https://github.com/thek4n/t/cmd/t@latest
```

### Download binary
```sh
wget https://github.com/TheK4n/t/releases/download/v0.1.2/t_0.1.2_linux_amd64.tar.gz
tar xzf t_0.1.2_linux_amd64.tar.gz
```


## Usage
```sh
t --help
t a Buy bread  # Add task
t   # Show tasks
t 1 # Show task with index 1
```