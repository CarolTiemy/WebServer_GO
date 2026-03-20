# ⚙️ WebServer GO

Webserver HTTP em Go puro focado em **performance máxima**. Dois endpoints: health check e compressão GZIP com tempo de resposta na casa dos **microsegundos**.

---

## 🚀 Quickstart

```bash
# Subir o servidor
make run

# Testar (em outro terminal)
make health
make process
make process JSON='{"sua":"msg"}'

# Parar
make stop
```

Ou acesse **http://localhost:8080** no navegador pra interface visual.

---

## 📡 Endpoints

### `GET /health`

Retorna o status do servidor.

```bash
$ curl http://localhost:8080/health
{"status":"ok","t":"1440ns"}
```

### `POST /process`

Recebe um JSON, comprime com GZIP e retorna os tamanhos.

```bash
$ curl -X POST http://localhost:8080/process \
  -H "Content-Type: application/json" \
  -d '{"msg":"hello"}'
{"original":15,"comprimido":40,"tempo":"646.24 us"}%                                                                                                                                               
```

| Campo | Descrição |
|---|---|
| `original` | Tamanho do body em bytes |
| `comprimido` | Tamanho após compressão GZIP |
| `tempo` | Tempo de processamento em nanosegundos |

---

## ⚡ Otimizações

| Técnica | O que faz | Impacto |
|---|---|---|
| **sync.Pool** | Reutiliza compressores e buffers entre requests | Elimina alocações por request |
| **klauspost/compress** | GZIP com instruções assembly (SSE/AVX) | ~2x mais rápido que stdlib |
| **nanotime()** | Clock monotônico direto do runtime (~10ns) | 7x mais barato que `time.Now()` |
| **Pre-warm** | Aquece 16 compressores no startup | Primeira request sem custo extra |
| **fmt.Fprintf direto** | Escreve JSON sem reflection | Zero alocação na resposta |
| **Medição cirúrgica** | Mede só a compressão, não I/O | Tempo reportado = processamento real |

### Resultado

| Cenário | V1 (ingênuo) | V4 (final) | Ganho |
|---|---|---|---|
| JSON pequeno (~30B) | 490 µs | **3-7 µs** | ~98% |
| JSON grande (~10KB) | 1.000 µs | **30 µs** | ~97% |

---

## 🏗️ Estrutura

```
.
├── main.go          # Servidor (handlers + pools + otimizações)
├── Makefile         # Comandos rápidos
├── go.mod
├── go.sum
└── static/
    ├── index.html   # Interface visual
    └── style.css    # Estilos
```

**Dependências externas:** apenas `github.com/klauspost/compress` (GZIP otimizado).

---

## 🛠️ Requisitos

- Go 1.21+

---

## 👥 Autores

- [CarolTiemy](https://github.com/CarolTiemy)
- [joao-paulino-nasc](https://github.com/joao-paulino-nasc)
