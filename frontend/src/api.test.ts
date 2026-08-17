import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, api, items, json, localDateTime } from './api'

afterEach(()=>{vi.unstubAllGlobals();vi.restoreAllMocks()})

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
    vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response('{"error":{"code":"VALIDATION_ERROR","message":"Faixa inválida","field":"network"}}',{status:422,headers:{'Content-Type':'application/json'}})))
    const error=await api('/scans').catch(value=>value as ApiError)
    expect(error).toBeInstanceOf(ApiError)
    expect(error.message).toBe('Faixa inválida')
    expect(error.field).toBe('network')
  })

  it('formata datetime-local respeitando o fuso do navegador',()=>{
    vi.spyOn(Date.prototype,'getTimezoneOffset').mockReturnValue(180)
    expect(localDateTime(new Date('2026-08-16T23:30:00.000Z'))).toBe('2026-08-16T20:30')
  })
})
