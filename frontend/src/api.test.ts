import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, items, json } from './api'

afterEach(()=>vi.unstubAllGlobals())

describe('api',()=>{
  it('normaliza listas e cria requisição JSON',()=>{
    expect(items<{id:string}>({})).toEqual([])
    expect(json('POST',{name:'Rack'})).toEqual({method:'POST',body:'{"name":"Rack"}'})
  })

  it('prefixa a versão e interpreta respostas',async()=>{
    const fetchMock=vi.fn().mockResolvedValue(new Response('{"status":"ok"}',{status:200,headers:{'Content-Type':'application/json'}}))
    vi.stubGlobal('fetch',fetchMock)
    await expect(api<{status:string}>('/system/status')).resolves.toEqual({status:'ok'})
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/system/status',expect.any(Object))
  })

  it('expõe a mensagem segura retornada pelo backend',async()=>{
    vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response('{"error":{"message":"Faixa inválida"}}',{status:422,headers:{'Content-Type':'application/json'}})))
    await expect(api('/scans')).rejects.toThrow('Faixa inválida')
  })
})
