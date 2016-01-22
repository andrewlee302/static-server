# A simple server for static html files

### Features

* *Cached file*, a file is visited faster after the first time because of loading to memory.
* *Dynamic loading*, when a static file changes, the server will reload it without restarting.
* Check and update the website periodically.
* Delete the cache entry if resouce is checked to be deleted.

### Usage

```
Args: [port] [rootDir]
Port should be a number, default value is 80.
RootDir should be an existed and absolute dir, default value is the working dir.
```

### TODO
1. Data consistency.
2. Checking all files periodically is too much work.