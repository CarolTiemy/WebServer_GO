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

func nanotime() int64

var poolCompressores = sync.Pool{
	New: func() any {
		compressor, _ := kgzip.NewWriterLevel(io.Discard, kgzip.BestSpeed)
		return compressor
	},
}

var poolBuffers = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func preWarm() {
	for i := 0; i < 16; i++ {
		buf := poolBuffers.Get().(*bytes.Buffer)
		buf.Reset()
		comp := poolCompressores.Get().(*kgzip.Writer)
		comp.Reset(buf)
		comp.Write([]byte(`{"warm":true}`))
		comp.Close()
		poolCompressores.Put(comp)
		poolBuffers.Put(buf)
	}
}

func healthHandler(resposta http.ResponseWriter, requisicao *http.Request) {
	inicio := nanotime()

	if requisicao.Method != http.MethodGet {
		http.Error(resposta, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resposta.Header().Set("Content-Type", "application/json")
	resposta.WriteHeader(http.StatusOK)
	fmt.Fprintf(resposta, `{"status":"ok","t":"%dns"}`, nanotime()-inicio)
}

func processHandler(resposta http.ResponseWriter, requisicao *http.Request) {
	conteudoOriginal, erro := io.ReadAll(requisicao.Body)
	if erro != nil {
		http.Error(resposta, "Failed to read body", http.StatusBadRequest)
		return
	}
	requisicao.Body.Close()

	if requisicao.Method != http.MethodPost {
		http.Error(resposta, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !json.Valid(conteudoOriginal) {
		http.Error(resposta, "Invalid JSON", http.StatusBadRequest)
		return
	}

	inicio := nanotime()

	buffer := poolBuffers.Get().(*bytes.Buffer)
	buffer.Reset()

	compressor := poolCompressores.Get().(*kgzip.Writer)
	compressor.Reset(buffer)

	compressor.Write(conteudoOriginal)
	compressor.Close()

	tamanhoComprimido := buffer.Len()
	duracao := nanotime() - inicio

	poolCompressores.Put(compressor)
	poolBuffers.Put(buffer)

	resposta.Header().Set("Content-Type", "application/json")
	resposta.WriteHeader(http.StatusOK)
	fmt.Fprintf(resposta,
		`{"original":%d,"comprimido":%d,"tempo":"%dns"}`,
		len(conteudoOriginal), tamanhoComprimido, duracao)
}

func main() {
	inicioStartup := time.Now()

	preWarm()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/process", processHandler)

	arquivosEstaticos := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", arquivosEstaticos))

	http.HandleFunc("/", func(resposta http.ResponseWriter, requisicao *http.Request) {
		if requisicao.URL.Path != "/" {
			http.NotFound(resposta, requisicao)
			return
		}
		http.ServeFile(resposta, requisicao, "static/index.html")
	})

	porta := ":8080"
	log.Printf("Servidor pronto em %s (pools pre-aquecidos)", time.Since(inicioStartup))
	log.Printf("Rodando em http://localhost%s", porta)
	if erro := http.ListenAndServe(porta, nil); erro != nil {
		log.Fatal(erro)
	}
}
