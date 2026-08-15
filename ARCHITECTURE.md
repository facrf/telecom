# Arquitetura

O Telecom é um monólito local: um binário Go serve a API REST (`/api/v1`), o healthcheck e os ativos React compilados. A persistência é SQLite no volume de dados.

```text
Navegador → React incorporado → net/http / chi → serviços e repositórios → SQLite
```

## Camadas

- `cmd/telecom`: composição e ciclo de vida do processo.
- `internal/api`: HTTP, serialização e cabeçalhos de segurança; sem lógica de domínio.
- `internal/database`: conexão SQLite e migrations embutidas.
- `internal/web`: frontend compilado incorporado com `embed.FS`.
- `internal/clients`: entidades, validações e repositório de clientes/projetos.
- `internal/devices`: categorias, inventário e endereços de equipamentos.
- `internal/scanner` e `internal/discovery`: worker pools, scans, progresso, providers e diff.
- `internal/fingerprint`: motor de evidências e base OUI local/importável.
- `internal/diagrams` e `internal/documents`: topologias e documentação estruturada.
- `internal/attachments`, `internal/backup` e `internal/transfer`: arquivos, recuperação e formatos portáveis.
- `internal/technicalvisits`: domínio, validação, serviço e persistência das visitas técnicas.
- `frontend/src/pages`: módulos React carregados sob demanda; React Flow no editor visual.

## Persistência

As migrations `001` a `014` cobrem configurações/auditoria, clientes/projetos, inventário, anexos/tags, scans de rede e portas, documentação, diagramas, eventos de diff, fingerprints, OUI e as fases VT-1/VT-2 das visitas técnicas. A tabela `schema_migrations` controla a evolução sem SQL manual.

## Visitas técnicas

`technical_visits` referencia somente `projects.id`; cliente e seus dados atuais são obtidos por `JOIN`, eliminando a possibilidade de uma visita apontar para cliente e projeto incompatíveis. A FK usa `ON DELETE RESTRICT`: excluir a visita nunca afeta projeto/cliente e um projeto com histórico não é apagado por cascata.

O service valida os valores de domínio e gera IDs internos aleatórios. O repository gera o protocolo amigável dentro da mesma transação do insert e usa `updated_at` como token de concorrência nos updates.

A VT-2 armazena equipamentos relacionados, serviços, checklist, materiais e pendências em tabelas dependentes, removidas em cascata somente quando a visita é excluída. Equipamentos são referências ao inventário existente. Validações no service/repository e triggers SQLite garantem que visita e equipamento pertençam ao mesmo projeto. Anexos, scans, diagramas e assinaturas permanecem para migrations posteriores.

SQLite usa `foreign_keys`, `busy_timeout` e WAL. Backups usam uma cópia consistente e restaurações são preparadas em staging antes da troca recuperável dos dados.

## Scanner

O scanner nativo valida o alvo, limita hosts e distribui trabalho entre um número fixo de workers. Cada host combina respostas de TCP, ICMP, ARP, DNS reverso e descoberta multicast. O fingerprint é uma etapa separada, pontuada por evidências. Eventos SSE não possuem timeout global de escrita; conexão e probes possuem limites próprios e respeitam cancelamento.
