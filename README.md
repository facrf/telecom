# Telecom

Aplicação web local e autocontida para técnicos de telecomunicações documentarem clientes, instalações e infraestrutura de rede. O projeto não usa telemetria, CDN ou serviços cloud obrigatórios.

## Funcionalidades

- dashboard, clientes, projetos e inventário de equipamentos;
- visitas técnicas por projeto, com protocolo anual, status, resultado, equipamentos, serviços, checklist, materiais e pendências;
- endereços, categorias, tags e anexos validados;
- scanner de rede com CIDR/faixa, worker pool, cancelamento e progresso SSE;
- descoberta por TCP, ICMP, ARP, reverse DNS, mDNS, SSDP e ONVIF;
- scan de portas rápido, padrão e personalizado, banners e fingerprinting;
- histórico e comparação persistente de scans;
- editor React Flow com topologias persistentes e exportação SVG, PNG e PDF;
- documentação estruturada e relatórios PNG/PDF;
- importação/exportação JSON versionada e transacional;
- backup SQLite consistente e restauração validada com rollback;
- pesquisa global, configurações, auditoria e modos claro/escuro.

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
cd frontend && npm run lint && npm run test && npm run build
```

O build copia a saída do Vite para `internal/web/static`, que é incorporada com `embed.FS`.

## Visitas técnicas

As FASES VT-1 e VT-2 estão disponíveis em **Visitas técnicas** no menu e pelo botão **Visitas** de cada projeto. O módulo mantém o vínculo normalizado pelo projeto, gera protocolos no formato `VT-AAAA-NNNNNN` e oferece CRUD, filtros, duração calculada, controle otimista de concorrência e registros estruturados do trabalho executado. Consulte [docs/technical-visits.md](docs/technical-visits.md) para a API, teste manual e escopo das próximas fases.

## Docker

Você pode utilizar a imagem pré-compilada do GitHub Container Registry (GHCR) ou realizar uma build local.

**Executando via GHCR:**

```bash
docker pull ghcr.io/facrf/telecom:latest
docker run -d --name telecom --network host -v telecom-data:/data ghcr.io/facrf/telecom:latest
```

*(Ou com build local: `docker build -t telecom:local .` e `docker run -d --name telecom --network host -v ./telecom-data:/data telecom:local`)*

> Com `docker run`, utilize `--network host` para que os recursos de varredura (ARP, mDNS, SSDP, ONVIF) alcancem a LAN do host. Caso não utilize rede host, publique a porta com `-p 14000:14000`.

## Instalação via Portainer

Para implantar a aplicação no Portainer utilizando o [portainer-stack.yml](portainer-stack.yml):

1. **Imagem no GHCR:**
   A stack utiliza a imagem oficial `ghcr.io/facrf/telecom:latest` (você também pode fixar uma tag específica, como `ghcr.io/facrf/telecom:sha-5e54651`).
   Opcionalmente, baixe a imagem no host com antecedência:
   ```bash
   docker pull ghcr.io/facrf/telecom:latest
   ```

2. **Criar a Stack:**
   - Acesse o Portainer e selecione o ambiente desejado (**Environments** > **local**).
   - No menu lateral, acesse **Stacks** e clique em **+ Add stack**.
   - No campo **Name**, informe `telecom`.
   - Em **Build method**, escolha **Web editor** e cole o conteúdo do [portainer-stack.yml](portainer-stack.yml):

   ```yaml
   services:
     telecom:
       image: ghcr.io/facrf/telecom:latest
       container_name: telecom
       restart: unless-stopped
       network_mode: host
       environment:
         TELECOM_PORT: "14000"
         TELECOM_DATA_DIR: /data
         TELECOM_SCAN_WORKERS: "32"
         TELECOM_LOG_LEVEL: info
       volumes:
         - telecom-data:/data
       read_only: true
       tmpfs:
         - /tmp
       security_opt:
         - no-new-privileges:true

   volumes:
     telecom-data:
       name: telecom-data
   ```

3. **Deploy:**
   - Clique em **Deploy the stack**. O Portainer fará o download da imagem e iniciará o container.
   - Após a inicialização, a aplicação estará disponível em `http://<IP_DO_HOST>:14000`.

> **Nota:** A stack utiliza `network_mode: host` para permitir que o scanner de rede descubra dispositivos via ARP, mDNS, SSDP e ONVIF na LAN local do host.

## Configuração

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `TELECOM_PORT` | `14000` | Porta HTTP |
| `TELECOM_DATA_DIR` | `./data` | Banco SQLite e anexos |
| `TELECOM_SCAN_WORKERS` | `32` | Concorrência máxima do scanner |
| `TELECOM_LOG_LEVEL` | `info` | `debug`, `info`, `warn` ou `error` |

## Persistência e backup

Os dados persistentes vivem no diretório configurado em `TELECOM_DATA_DIR`: `telecom.sqlite`, `attachments/` e `backups/`. O backup usa o mecanismo seguro `VACUUM INTO`. A restauração exige confirmação, valida manifest/ZIP e `PRAGMA integrity_check`, preserva o estado anterior e é aplicada durante reinicialização controlada.

## Segurança e permissões

Consulte [SECURITY.md](SECURITY.md). O scanner aceita redes privadas por padrão e deve ser usado somente com autorização.

## Licença

LICENSE TBD — a licença será definida pelo mantenedor.
