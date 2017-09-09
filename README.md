Xword Server
============

This is a web server written in Go which provides support for saving
and syncing crossword puzzles.  The DB schema is modeled after the
XPF (XML Puzzle Format) but other formats are supported.

The corresponding client, also by the author, is XwordJS, a
javascript/html5 crossword application, which you can find here:

    https://github.com/bcopeland/xwordjs

You can also try it out here:

    https://bobcopeland.com/crosswords/xwordjs/

Building
--------

```
$ export GOPATH=~/gocode
$ [go get....]
$ go build src/server.go
```


