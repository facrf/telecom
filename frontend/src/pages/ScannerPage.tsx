import { FormEvent, useEffect, useRef, useState } from 'react'
import { api, items, json } from '../api'
import { Project, Scan, ScanChange, ScanHost } from '../types'
import { Empty, Notice, PageHeader, Panel, Status } from '../components'

type Progress = {
  scanId: string
  status: string
  hostsScanned: number
  hostsFound: number
  total: number
  host?: ScanHost
}

export default function ScannerPage({initialProjectId='' }:{initialProjectId?:string}) {
  const [projects, setProjects] = useState<Project[]>([])
  const [project, setProject] = useState(initialProjectId)
  const [network, setNetwork] = useState('192.168.0.0/24')
  const [scans, setScans] = useState<Scan[]>([])
  const [selected, setSelected] = useState<Scan>()
  const [hosts, setHosts] = useState<ScanHost[]>([])
  const [changes, setChanges] = useState<ScanChange[]>([])
  const [progress, setProgress] = useState<Progress>()
  const [message, setMessage] = useState('')
  const [assignProjectId, setAssignProjectId] = useState('')
  const [selectedHosts, setSelectedHosts] = useState<Set<string>>(new Set())
  const stream = useRef<EventSource | undefined>(undefined)

  const loadProjects = async () => {
    try {
      const data = await api<{ items: Project[] }>('/projects')
      setProjects(items(data))
    } catch {
      setProjects([])
    }
  }

  const loadScans = async () => {
    try {
      const query = project ? `?project_id=${encodeURIComponent(project)}` : ''
      const data = await api<{ items: Scan[] }>(`/scans${query}`)
      setScans(items(data))
    } catch {
      setScans([])
    }
  }

  const loadResult = async (scan: Scan) => {
    setSelected(scan)
    setAssignProjectId(scan.projectId || '')
    try {
      const [hostValue, changeValue, fresh] = await Promise.all([
        api<{ items: ScanHost[] }>(`/scans/${scan.id}/hosts`),
        api<{ items: ScanChange[] }>(`/scans/${scan.id}/changes`),
        api<Scan>(`/scans/${scan.id}`)
      ])
      setHosts(items(hostValue))
      setSelectedHosts(new Set())
      setChanges(items(changeValue))
      setSelected(fresh)
      setAssignProjectId(fresh.projectId || '')
    } catch (err: unknown) {
      setMessage(`Erro ao carregar detalhes: ${(err as Error).message || err}`)
    }
  }

  useEffect(() => {
    void loadProjects()
    return () => stream.current?.close()
  }, [])

  useEffect(() => {
    void loadScans()
  }, [project])

  useEffect(() => setProject(initialProjectId), [initialProjectId])

  const start = async (event: FormEvent) => {
    event.preventDefault()
    try {
      const scan = await api<Scan>('/scans', json('POST', { projectId: project, network }))
      setSelected(scan)
      setAssignProjectId(scan.projectId || '')
      setHosts([])
      setSelectedHosts(new Set())
      setChanges([])
      setProgress({ scanId: scan.id, status: 'queued', hostsScanned: 0, hostsFound: 0, total: 0 })
      await loadScans()

      stream.current?.close()
      const source = new EventSource(`/api/v1/scans/${scan.id}/events`)
      stream.current = source

      source.addEventListener('progress', raw => {
        const value = JSON.parse((raw as MessageEvent).data) as Progress
        setProgress(value)
        if (value.host?.status === 'online') {
          setHosts(current => current.some(item => item.ip === value.host?.ip) ? current : [...current, value.host!])
        }
        if (value.status !== 'running' && value.status !== 'queued') {
          source.close()
          void loadScans()
          void loadResult({ ...scan, status: value.status })
        }
      })

      source.onerror = () => {
        source.close()
        void loadScans()
      }
    } catch (err: unknown) {
      setMessage(`Erro ao iniciar scan: ${(err as Error).message || err}`)
    }
  }

  const cancel = async () => {
    if (!selected) return
    try {
      await api(`/scans/${selected.id}/cancel`, { method: 'POST' })
      setMessage('Cancelamento solicitado.')
    } catch (err: unknown) {
      setMessage(`Erro ao cancelar: ${(err as Error).message || err}`)
    }
  }

  const linkProject = async (targetProjectId: string) => {
    if (!selected) return
    try {
      const updated = await api<Scan>(`/scans/${selected.id}`, json('PATCH', { projectId: targetProjectId }))
      setSelected(updated)
      setAssignProjectId(updated.projectId || '')
      await loadScans()
      setMessage(targetProjectId ? 'Scan vinculado ao projeto com sucesso.' : 'Vínculo do scan removido.')
    } catch (err: unknown) {
      setMessage(`Erro ao vincular projeto: ${(err as Error).message || err}`)
    }
  }

  const inventoryHosts = async (selectedItems:ScanHost[]) => {
    if (!selected || selectedItems.length === 0) return
    const targetProjectId = selected.projectId || assignProjectId || (projects.length === 1 ? projects[0].id : '')
    if (!targetProjectId) { setMessage('Selecione o projeto de destino antes de inventariar.'); return }
    try {
      const result=await api<{created:number;skipped:number}>(`/scans/${selected.id}/inventory`,json('POST',{projectId:targetProjectId,items:selectedItems.map(host=>({ip:host.ip,name:host.hostname||`${host.deviceType||'Dispositivo'} ${host.ip}`,categoryId:host.categoryId||'other'}))}))
      setSelectedHosts(new Set())
      setSelected(current=>current?{...current,projectId:targetProjectId}:current)
      setAssignProjectId(targetProjectId)
      await loadScans()
      setMessage(`${result.created} equipamento(s) inventariado(s). ${result.skipped?`${result.skipped} duplicado(s) ignorado(s).`:''}`)
    } catch (err: unknown) { setMessage(`Erro ao inventariar: ${(err as Error).message || err}`) }
  }
  const toggleHost=(ip:string)=>setSelectedHosts(current=>{const next=new Set(current);if(next.has(ip))next.delete(ip);else next.add(ip);return next})

  const percentage = progress?.total ? Math.round((progress.hostsScanned / progress.total) * 100) : 0
  const selectedProjectName = projects.find(p => p.id === selected?.projectId)?.name

  return (
    <>
      <PageHeader
        eyebrow="DESCOBERTA"
        title="Scanner de rede"
        description="Descoberta controlada em redes autorizadas com progresso em tempo real."
      />
      <div className="warning">
        ⚠ Utilize o scanner somente em redes e equipamentos para os quais você possui autorização.
      </div>
      <Notice message={message} onClose={() => setMessage('')} />

      <Panel title="Novo scan">
        <form className="scan-form" onSubmit={start}>
          <select value={project} onChange={e => setProject(e.target.value)}>
            <option value="">Sem projeto (Descoberta avulsa / Geral)</option>
            {projects.map(value => (
              <option key={value.id} value={value.id}>{value.name}</option>
            ))}
          </select>
          <input
            required
            value={network}
            onChange={e => setNetwork(e.target.value)}
            placeholder="192.168.0.0/24, 192.168.0.1-192.168.0.254 ou 192.168.0.1"
          />
          <button className="primary">Iniciar descoberta</button>
        </form>
        {progress && ['queued', 'running'].includes(progress.status) && (
          <div className="progress-card">
            <div>
              <b>Escaneando {network}</b>
              <Status value={progress.status} />
            </div>
            <progress max={100} value={percentage} />
            <div>
              <span>{progress.hostsScanned} / {progress.total} hosts</span>
              <span>Encontrados: {progress.hostsFound}</span>
              <button className="danger" onClick={() => void cancel()}>Cancelar scan</button>
            </div>
          </div>
        )}
      </Panel>

      <div className="split scanner-split">
        <Panel title="Histórico">
          <div className="select-list">
            {scans.length === 0 ? (
              <Empty>Nenhum scan realizado.</Empty>
            ) : (
              scans.map(scan => (
                <button
                  className={selected?.id === scan.id ? 'selected' : ''}
                  key={scan.id}
                  onClick={() => void loadResult(scan)}
                >
                  <div>
                    <b>{scan.network}</b>
                    <Status value={scan.status} />
                  </div>
                  <span>
                    {scan.hostsFound} encontrados · {scan.hostsScanned} verificados
                    {scan.projectId && ` · ${projects.find(p => p.id === scan.projectId)?.name || 'Projeto'}`}
                  </span>
                </button>
              ))
            )}
          </div>
        </Panel>

        <Panel title={selected ? `Resultados · ${selected.network}` : 'Resultados'}>
          {!selected ? (
            <Empty>Selecione ou inicie um scan.</Empty>
          ) : (
            <>
              <div className="tabs-summary" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '8px' }}>
                <div>
                  <span><b>{hosts.length}</b> hosts encontrados</span>
                  <span style={{ marginLeft: '12px' }}><b>{changes.length}</b> alterações</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <small style={{ color: 'var(--muted, #64748b)' }}>
                    Projeto: {selectedProjectName || 'Descoberta avulsa'}
                  </small>
                  <select
                    value={assignProjectId}
                    onChange={e => setAssignProjectId(e.target.value)}
                    style={{ fontSize: '0.85rem', padding: '2px 6px' }}
                  >
                    <option value="">Sem projeto</option>
                    {projects.map(p => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                  {assignProjectId !== (selected.projectId || '') && (
                    <button
                      className="primary"
                      style={{ fontSize: '0.8rem', padding: '3px 8px' }}
                      onClick={() => void linkProject(assignProjectId)}
                    >
                      Salvar vínculo
                    </button>
                  )}
                </div>
              </div>

              {hosts.length > 0 && (
                <div className="table-scroll responsive-table">
                  <div className="bulk-actions"><label><input type="checkbox" checked={selectedHosts.size===hosts.length} onChange={event=>setSelectedHosts(event.target.checked?new Set(hosts.map(host=>host.ip)):new Set())}/> Selecionar todos</label><button className="primary" disabled={selectedHosts.size===0} onClick={()=>void inventoryHosts(hosts.filter(host=>selectedHosts.has(host.ip)))}>Inventariar selecionados ({selectedHosts.size})</button></div>
                  <table>
                    <thead>
                      <tr>
                        <th></th>
                        <th>IP / hostname</th>
                        <th>Identificação</th>
                        <th>Métodos</th>
                        <th>Confiança</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {hosts.map(host => (
                        <tr key={host.ip}>
                          <td data-label="Selecionar"><input type="checkbox" aria-label={`Selecionar ${host.ip}`} checked={selectedHosts.has(host.ip)} onChange={()=>toggleHost(host.ip)}/></td>
                          <td data-label="IP / hostname">
                            <code>{host.ip}</code>
                            {host.mac && <small style={{ display: 'block', color: 'var(--muted, #64748b)' }}>MAC: {host.mac}</small>}
                            {host.hostname && <small style={{ display: 'block' }}>{host.hostname}</small>}
                          </td>
                          <td data-label="Identificação">
                            {host.deviceType || 'Dispositivo'}
                            {host.manufacturer && <small style={{ display: 'block' }}>{host.manufacturer}</small>}
                          </td>
                          <td data-label="Métodos">{(host.discoveryMethods || []).join(', ')}</td>
                          <td data-label="Confiança">
                            {Math.round(host.confidence * 100)}%
                            {host.evidence && host.evidence.length > 0 && (
                              <details>
                                <summary>Evidências</summary>
                                {host.evidence.map((evidence, index) => (
                                  <p key={index}>{evidence.source}: {evidence.detail}</p>
                                ))}
                              </details>
                            )}
                          </td>
                          <td className="row-actions">
                            <button onClick={() => void inventoryHosts([host])}>Inventariar</button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {changes.length > 0 && (
                <section>
                  <h3>Alterações desde o scan anterior</h3>
                  <div className="change-list">
                    {changes.map((change, index) => (
                      <div key={index}>
                        <b>{change.type || change.Type}</b>
                        <span>{change.subject || change.Subject}</span>
                        <small>{change.previous || change.Previous} → {change.current || change.Current}</small>
                      </div>
                    ))}
                  </div>
                </section>
              )}
            </>
          )}
        </Panel>
      </div>
    </>
  )
}
