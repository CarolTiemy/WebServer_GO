.PHONY: run build health process stop clean

# Roda o servidor
run:
	@cd $(dir $(abspath $(lastword $(MAKEFILE_LIST)))) && go run main.go

# Compila o binário
build:
	@go build -o server .
	@echo "Binário 'server' criado!"

# Testa o GET /health
health:
	@curl -s http://localhost:8080/health | jq .

# Testa o POST /process (uso: make process ou make process JSON='{"sua":"msg"}')
JSON ?= {"a":1}
process:
	@curl -s -X POST http://localhost:8080/process \
		-H "Content-Type: application/json" \
		-d '$(JSON)' | jq .

# Mata o servidor rodando na porta 8080
stop:
	@lsof -ti :8080 | xargs kill -9 2>/dev/null && echo "Servidor parado" || echo "Nenhum servidor rodando"

# Remove o binário
clean:
	@rm -f server
