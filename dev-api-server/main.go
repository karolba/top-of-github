package main

import (
	"log"
	"net/http"
)

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s\n", req.Method, req.URL)
		next.ServeHTTP(writer, req)
	})
}

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Origin", "*")

		// TODO: files in *contentDirectory are always gzip-compressed.
		// in production Cloudflare R2 makes sure to transparently uncompress them
		// if a user agent doesn't include the `Accept: gzip` header. This doesn't
		// happen here, but since this code is only helpful for offline development,
		// it's not really a concern.
		writer.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(writer, req)
	})
}

func middleware(h http.Handler) http.Handler {
	h = logRequest(h)
	h = setHeaders(h)
	return h
}

func main() {
	handler := middleware(http.FileServer(http.Dir(*contentDirectory)))
	err := http.ListenAndServe(*listenAddress, handler)
	if err != nil {
		log.Fatalln(err)
	}
}
