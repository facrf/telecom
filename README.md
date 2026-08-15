# Telecom

Aplicação web local e autocontida para técnicos de telecomunicações documentarem clientes, instalações e infraestrutura de rede. O projeto não usa telemetria, CDN ou serviços cloud obrigatórios.

## Estado atual

As Fases 1 e 2 estão implementadas: fundação do servidor local e CRUD de clientes/projetos com pesquisa.

## Desenvolvimento

Requer Go 1.24+ e Node.js 22+.

```bash
cd frontend && npm install && npm run dev
```

Em outro terminal:

```bash
go run ./cmd/telecom
```

Abra `http://localhost:5173`. A API está em `http://localhost:14000`; `GET /health` responde com o status do serviço.

## Build e testes

```bash
make build
make test
make vet
```

O build copia a saída do Vite para `internal/web/static`, que é incorporada com `embed.FS`.

## Docker

```bash
docker build -t telecom:local .
docker run -d --name telecom -p 14000:14000 -v ./telecom-data:/data telecom:local
```

## Configuração

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `TELECOM_PORT` | `14000` | Porta HTTP |
| `TELECOM_DATA_DIR` | `./data` | Banco SQLite e anexos |
| `TELECOM_SCAN_WORKERS` | `32` | Reserva para o scanner da Fase 4 |
| `TELECOM_LOG_LEVEL` | `info` | `debug`, `info`, `warn` ou `error` |

## Persistência e backup

Os dados persistentes vivem no diretório configurado em `TELECOM_DATA_DIR`. O backup/restauração consistente será introduzido na Fase 13.

## Segurança e permissões

Consulte [SECURITY.md](SECURITY.md). Recursos de scanner serão adicionados somente para redes sob autorização do operador.

## Licença

LICENSE TBD — a licença será definida pelo mantenedor.
