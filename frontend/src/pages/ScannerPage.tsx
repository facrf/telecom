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

export default function ScannerPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [project, setProject] = useState('')
  const [network, setNetwork] = useState('192.168.0.0/24')
  const [scans, setScans] = useState<Scan[]>([])
  const [selected, setSelected] = useState<Scan>()
  const [hosts, setHosts] = useState<ScanHost[]>([])
  const [changes, setChanges] = useState<ScanChange[]>([])
  const [progress, setProgress] = useState<Progress>()
  const [message, setMessage] = useState('')
  const [assignProjectId, setAssignProjectId] = useState('')
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

  const start = async (event: FormEvent) => {
    event.preventDefault()
    try {
      const scan = await api<Scan>('/scans', json('POST', { projectId: project, network }))
      setSelected(scan)
      setAssignProjectId(scan.projectId || '')
      setHosts([])
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

  const addDevice = async (host: ScanHost) => {
    if (!selected) return

    let targetProjectId = selected.projectId
    if (!targetProjectId) {
      if (projects.length === 0) {
        setMessage('Nenhum projeto cadastrado no sistema. Cadastre um cliente e projeto no menu "Clientes e projetos" antes de inventariar.')
        return
      }
      if (projects.length === 1) {
        targetProjectId = projects[0].id
      } else {
        const optionsText = projects.map((p, i) => `${i + 1}. ${p.name}`).join('\n')
        const chosen = prompt(`Selecione o projeto para o equipamento:\n${optionsText}\n\nDigite o número ou ID do projeto:`, '1')
        if (!chosen) return
        const index = parseInt(chosen, 10) - 1
        if (!isNaN(index) && projects[index]) {
          targetProjectId = projects[index].id
        } else {
          const matched = projects.find(p => p.id === chosen || p.name.toLowerCase() === chosen.toLowerCase())
          if (matched) {
            targetProjectId = matched.id
          } else {
            setMessage('Projeto selecionado inválido.')
            return
          }
        }
      }
    }

    const defaultName = host.hostname || `${host.deviceType || 'Dispositivo'} ${host.ip}`
    const name = prompt('Nome do equipamento para o inventário:', defaultName)
    if (!name) return

    try {
      const device = await api<{ id: string }>('/devices', json('POST', {
        projectId: targetProjectId,
        name,
        categoryId: 'other',
        manufacturer: host.manufacturer,
        hostname: host.hostname,
        status: 'online'
      }))

      await api(`/devices/${device.id}/addresses`, json('POST', {
        type: 'ipv4',
        address: host.ip,
        primary: true
      }))

      if (host.mac) {
        await api(`/devices/${device.id}/addresses`, json('POST', {
          type: 'mac',
          address: host.mac,
          primary: true
        }))
      }

      if (!selected.projectId && targetProjectId) {
        await linkProject(targetProjectId)
      }

      const projectName = projects.find(p => p.id === targetProjectId)?.name || 'projeto'
      setMessage(`Equipamento "${name}" adicionado com sucesso ao ${projectName}.`)
    } catch (err: unknown) {
      setMessage(`Erro ao inventariar equipamento: ${(err as Error).message || err}`)
    }
  }

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
                <div className="table-scroll">
                  <table>
                    <thead>
                      <tr>
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
                          <td>
                            <code>{host.ip}</code>
                            {host.mac && <small style={{ display: 'block', color: 'var(--muted, #64748b)' }}>MAC: {host.mac}</small>}
                            {host.hostname && <small style={{ display: 'block' }}>{host.hostname}</small>}
                          </td>
                          <td>
                            {host.deviceType || 'Dispositivo'}
                            {host.manufacturer && <small style={{ display: 'block' }}>{host.manufacturer}</small>}
                          </td>
                          <td>{(host.discoveryMethods || []).join(', ')}</td>
                          <td>
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
                          <td>
                            <button onClick={() => void addDevice(host)}>Inventariar</button>
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
