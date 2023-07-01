package main

import "flag"

var contentDirectory = flag.String("dir", ".", "from where to serve files")
var listenAddress = flag.String("listen-on", "127.0.0.1:10002", "address to listen on")

func init() {
	flag.Parse()
}
