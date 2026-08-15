# Visitas técnicas — FASES VT-1 e VT-2

## Escopo entregue

A primeira fase implementa cadastro, consulta, edição e exclusão de visitas, sempre vinculadas a um projeto existente. A resposta da API inclui cliente e projeto resolvidos pelo relacionamento normalizado. Também estão disponíveis protocolo automático, status, resultado, campos básicos do relato, cálculo de duração, filtros e pesquisa global.

A segunda fase acrescenta equipamentos envolvidos por referência ao inventário, serviços executados, checklist ordenável, materiais utilizados e pendências priorizadas. Todos possuem criação, listagem, edição e exclusão próprias.

Status aceitos: `draft`, `scheduled`, `in_progress`, `completed` e `cancelled`.

Resultados aceitos: `resolved`, `partially_resolved`, `not_resolved`, `waiting_material`, `waiting_customer`, `requires_return` e `no_fault_found`. O resultado pode ficar vazio enquanto a visita não estiver concluída.

## API

| Método | Endpoint | Finalidade |
| --- | --- | --- |
| `GET` | `/api/v1/technical-visits` | Lista e filtra visitas |
| `POST` | `/api/v1/technical-visits` | Cria visita e protocolo |
| `GET` | `/api/v1/technical-visits/{id}` | Obtém uma visita |
| `PUT` | `/api/v1/technical-visits/{id}` | Atualiza usando `updatedAt` |
| `DELETE` | `/api/v1/technical-visits/{id}` | Exclui a visita |
| `GET` | `/api/v1/projects/{id}/technical-visits` | Lista visitas do projeto |
| `GET` | `/api/v1/clients/{id}/technical-visits` | Lista visitas dos projetos do cliente |

Cada grupo VT-2 possui `GET` e `POST` na coleção, além de `PUT` e `DELETE` em `/{itemID}`:

| Recurso | Coleção |
| --- | --- |
| Equipamentos | `/api/v1/technical-visits/{id}/devices` |
| Serviços | `/api/v1/technical-visits/{id}/services` |
| Checklist | `/api/v1/technical-visits/{id}/checklist` |
| Materiais | `/api/v1/technical-visits/{id}/materials` |
| Pendências | `/api/v1/technical-visits/{id}/pending-items` |

Um equipamento somente pode ser relacionado à visita ou a um serviço quando pertence ao mesmo projeto. A regra existe tanto no código quanto em triggers SQLite. Excluir a visita remove esses relacionamentos e dados exclusivos, mas nunca remove o equipamento do inventário.

Filtros da listagem: `q`, `client_id`, `project_id`, `technician`, `status`, `result`, `date_from` e `date_to`.

Exemplo mínimo:

```json
{
  "projectId": "id-do-projeto",
  "title": "Diagnóstico do link principal",
  "visitType": "diagnosis",
  "status": "draft",
  "scheduledAt": "2026-08-15T13:30"
}
```

O `protocol` é somente um identificador amigável; a chave estável é `id`. Em edição, o cliente deve enviar o `updatedAt` recebido. Uma versão antiga retorna HTTP `409` com o código `CONCURRENT_UPDATE`.

## Teste manual

1. Cadastre um cliente e um projeto em **Clientes e projetos**.
2. Clique em **Visitas** na linha do projeto.
3. Crie uma visita e confira o protocolo `VT-AAAA-NNNNNN`.
4. Edite as etapas básicas do formulário e salve.
5. Reabra a visita e registre equipamentos, serviços, checklist, materiais e pendências.
6. Confirme que o seletor de equipamentos mostra somente o inventário do projeto.
7. Aplique filtros por cliente, projeto, status, resultado e período.
8. Pesquise o protocolo, um serviço ou equipamento relacionado na busca global.
9. Abra e exclua a visita, confirmando o diálogo.
10. Para concorrência, abra a mesma visita em duas sessões, salve na primeira e tente salvar a versão antiga na segunda; a API deve responder `409`.

## Próxima fase (VT-3)

Permanecem fora desta fase: anexos, fotos, categorias de fotos e pares antes/depois. A VT-3 deverá reutilizar o mecanismo polimórfico de attachments existente, sem criar outro sistema de upload.
