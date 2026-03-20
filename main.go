package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	_ "unsafe"

	kgzip "github.com/klauspost/compress/gzip"
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

var compressores = sync.Pool{New: func() any { w, _ := kgzip.NewWriterLevel(io.Discard, kgzip.BestSpeed); return w }}
var buffers = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func formatNs(ns int64) string {
	switch {
	case ns < 1000:
		return fmt.Sprintf("%d ns", ns)
	case ns < 1000000:
		return fmt.Sprintf("%.2f us", float64(ns)/1000)
	case ns < 1000000000:
		return fmt.Sprintf("%.2f ms", float64(ns)/1000000)
	default:
		return fmt.Sprintf("%.2f s", float64(ns)/1000000000)
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	t := nanotime()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","t":"%s"}`, formatNs(nanotime()-t))
}

func process(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil || !json.Valid(body) {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	t := nanotime()

	buf := buffers.Get().(*bytes.Buffer)
	buf.Reset()
	gz := compressores.Get().(*kgzip.Writer)
	gz.Reset(buf)
	gz.Write(body)
	gz.Close()

	comp := buf.Len()
	dur := nanotime() - t

	compressores.Put(gz)
	buffers.Put(buf)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"original":%d,"comprimido":%d,"tempo":"%s"}`, len(body), comp, formatNs(dur))
}

func main() {
	t := time.Now()

	// Pre-warm: aquece os pools pra primeira request não ser lenta
	for i := 0; i < 16; i++ {
		buf := buffers.Get().(*bytes.Buffer)
		buf.Reset()
		gz := compressores.Get().(*kgzip.Writer)
		gz.Reset(buf)
		gz.Write([]byte(`{}`))
		gz.Close()
		compressores.Put(gz)
		buffers.Put(buf)
	}

	http.HandleFunc("/health", health)
	http.HandleFunc("/process", process)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	log.Printf("Pronto em %s → http://localhost:8080", time.Since(t))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
