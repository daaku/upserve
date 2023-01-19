package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const indexHTML = `<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="x-ua-compatible" content="ie=edge">
<title>Upload</title>
<style>
main { display: grid; height: 100vh; }
form { place-self: center; }
</style>
<main>
	<form enctype="multipart/form-data" action="/upload" method="post">
		<input type="file" name="file" multiple>
		<button type="submit">Upload</button>
	</form>
</main>
`

const successHTML = `<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="x-ua-compatible" content="ie=edge">
<meta http-equiv="refresh" content="10;url=/">
<title>Success</title>
<style>
main { display: grid; height: 100vh; }
div { place-self: center; }
</style>
<main>
	<div>
		<p>Upload successful.</p>
		<small>Redirecting to home in 10s.</small>
	</div>
</main>
`

func sendError(w http.ResponseWriter, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintln(w, err)
}

func main() {
	wd, _ := os.Getwd()
	dir := flag.String("dir", filepath.Join(wd, "incoming"), "directory to save files to")
	addr := flag.String("addr", "0.0.0.0:8080", "address to bind to")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, indexHTML)
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		mpr, err := r.MultipartReader()
		if err != nil {
			sendError(w, err)
			return
		}
		for {
			part, err := mpr.NextPart()
			if err == io.EOF {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				io.WriteString(w, successHTML)
				return
			}
			if err != nil {
				sendError(w, err)
				return
			}
			if err := os.MkdirAll(*dir, 0755); err != nil {
				sendError(w, err)
				return
			}
			dstPath := filepath.Join(*dir, part.FileName())
			dst, err := os.Create(dstPath)
			if err != nil {
				sendError(w, err)
				return
			}
			if _, err := io.Copy(dst, part); err != nil {
				sendError(w, err)
				return
			}
			if err := dst.Close(); err != nil {
				sendError(w, err)
				return
			}
			if err := part.Close(); err != nil {
				sendError(w, err)
				return
			}
			log.Println("Recieved file", dstPath)
		}
	})
	log.Printf("Serving at http://%s/ uploading to %s\n", *addr, *dir)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
