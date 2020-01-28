# Handy Handlers

Collection of useful HTTP handlers that form a handler stack of filters.  Many
are opinionated for a specific purpose.

[![GoDoc](https://godoc.org/github.com/streadway/handy?status.svg)](https://godoc.org/github.com/streadway/handy)

# API

The signature for a handler wrapper ends with an http.Handler interface and returns a http.Handler interface.  For example:

```go
func Log(w io.Writer, http.Handler) http.Handler

func CORS(methods, origins []string, http.Handler) http.Handler
```

# Contributing

Fork the repo and add your fix or handler with tests.  Make a pull request.

