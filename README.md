# A simple server for static html files

### Features

* *Cached file*, a file is visited faster at the second time because of loading to memory.
* *Dynamic loading*, when a static file changes, the server will reload it without restarting.

### Usage

```
Args: [port] [rootDir]
Port should be a number, default value is 80
RootDir should be an existed and absolute dir, default value is the working dir
```