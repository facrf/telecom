import { describe, expect, it } from 'vitest'
import { resultLabels, visitTypeLabels } from './VisitsPage'

describe('visitas técnicas',()=>{
  it('expõe os tipos e resultados iniciais em português',()=>{
    expect(Object.keys(visitTypeLabels)).toHaveLength(16)
    expect(visitTypeLabels.corrective_maintenance).toBe('Manutenção corretiva')
    expect(resultLabels.requires_return).toBe('Necessita retorno')
  })
})
