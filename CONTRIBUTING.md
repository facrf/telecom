# Contribuindo

Mantenha as alterações pequenas, cobertas por testes e organizadas por módulo. Toda evolução do banco requer migration nova; migrations existentes não devem ser alteradas após distribuição.

Antes de enviar uma alteração, execute:

```bash
make fmt
make test
make vet
cd frontend && npm run lint && npm run test && npm run build
```
