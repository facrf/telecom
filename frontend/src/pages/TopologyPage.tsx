import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  Background,
  Connection,
  Controls,
  Edge,
  Handle,
  MiniMap,
  Node,
  NodeChange,
  NodeProps,
  Position,
  ReactFlow,
  ReactFlowProvider,
  addEdge,
  applyEdgeChanges,
  applyNodeChanges,
  getNodesBounds,
  getViewportForBounds,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { toPng } from 'html-to-image'
import { jsPDF } from 'jspdf'
import { api, items, json } from '../api'
import { Category, Device, Diagram, DiagramEdge, DiagramNode, Project, Scan, ScanHost } from '../types'
import { ConfirmButton, Empty, Field, Modal, Notice, PageHeader, Panel, Status } from '../components'

type Graph = { diagram: Diagram; nodes: DiagramNode[]; edges: DiagramEdge[] }

export type NodeData = {
  label: string
  deviceId?: string
  device?: Device
  category?: string
  icon?: string
  color?: string
  ip?: string
  mac?: string
  vlan?: string
  notes?: string
}

const CATEGORY_META: Record<string, { label: string; icon: string; color: string }> = {
  gateway: { label: 'WAN / Internet', icon: '🌐', color: '#ef4444' },
  router: { label: 'Roteador', icon: '🔀', color: '#3b82f6' },
  firewall: { label: 'Firewall / UTM', icon: '🛡️', color: '#f43f5e' },
  switch: { label: 'Switch de Rede', icon: '🔲', color: '#06b6d4' },
  'access-point': { label: 'Access Point Wi-Fi', icon: '📶', color: '#8b5cf6' },
  'ip-camera': { label: 'Câmera IP / CFTV', icon: '📹', color: '#10b981' },
  nvr: { label: 'NVR / Gravador', icon: '📼', color: '#14b8a6' },
  dvr: { label: 'DVR CFTV', icon: '📼', color: '#14b8a6' },
  olt: { label: 'OLT GPON', icon: '🖧', color: '#a855f7' },
  onu: { label: 'ONU / ONT Fibra', icon: '🏠', color: '#d946ef' },
  ont: { label: 'ONT Fibra', icon: '🏠', color: '#d946ef' },
  server: { label: 'Servidor CPD', icon: '🖥️', color: '#f97316' },
  nas: { label: 'Storage / NAS', icon: '💾', color: '#eab308' },
  storage: { label: 'Storage', icon: '💾', color: '#eab308' },
  ups: { label: 'Nobreak / UPS', icon: '🔋', color: '#eab308' },
  printer: { label: 'Impressora', icon: '🖨️', color: '#64748b' },
  'ip-phone': { label: 'Telefone IP', icon: '📞', color: '#06b6d4' },
  ata: { label: 'Adaptador ATA', icon: '📞', color: '#06b6d4' },
  radio: { label: 'Rádio PTP', icon: '📡', color: '#8b5cf6' },
  antenna: { label: 'Antena Wireless', icon: '📡', color: '#8b5cf6' },
  'patch-panel': { label: 'Patch Panel', icon: '▦', color: '#64748b' },
  desktop: { label: 'Estação / PC', icon: '💻', color: '#3b82f6' },
  notebook: { label: 'Notebook', icon: '💻', color: '#3b82f6' },
  iot: { label: 'Dispositivo IoT', icon: '⚙', color: '#10b981' },
  other: { label: 'Equipamento', icon: '◈', color: '#64748b' },
}

const TEMPLATES = [
  { key: 'wan', label: 'Internet / WAN', category: 'gateway', icon: '🌐', color: '#ef4444' },
  { key: 'router', label: 'Roteador Borda', category: 'router', icon: '🔀', color: '#3b82f6' },
  { key: 'firewall', label: 'Firewall UTM', category: 'firewall', icon: '🛡️', color: '#f43f5e' },
  { key: 'switch-core', label: 'Switch Core', category: 'switch', icon: '🔲', color: '#06b6d4' },
  { key: 'switch-poe', label: 'Switch PoE', category: 'switch', icon: '🔌', color: '#0ea5e9' },
  { key: 'ap', label: 'AP Wi-Fi', category: 'access-point', icon: '📶', color: '#8b5cf6' },
  { key: 'camera', label: 'Câmera IP', category: 'ip-camera', icon: '📹', color: '#10b981' },
  { key: 'nvr', label: 'NVR CFTV', category: 'nvr', icon: '📼', color: '#14b8a6' },
  { key: 'olt', label: 'OLT GPON', category: 'olt', icon: '🖧', color: '#a855f7' },
  { key: 'onu', label: 'ONU / ONT', category: 'onu', icon: '🏠', color: '#d946ef' },
  { key: 'server', label: 'Servidor CPD', category: 'server', icon: '🖥️', color: '#f97316' },
  { key: 'nas', label: 'Storage NAS', category: 'nas', icon: '💾', color: '#eab308' },
  { key: 'ups', label: 'Nobreak UPS', category: 'ups', icon: '🔋', color: '#eab308' },
  { key: 'radio', label: 'Rádio PTP', category: 'radio', icon: '📡', color: '#8b5cf6' },
  { key: 'phone', label: 'Telefone IP', category: 'ip-phone', icon: '📞', color: '#06b6d4' },
  { key: 'patch', label: 'Patch Panel', category: 'patch-panel', icon: '▦', color: '#64748b' },
]

function TelecomNodeComponent({ data, selected }: NodeProps) {
  const nodeData = data as unknown as NodeData
  const cat = CATEGORY_META[nodeData.category || 'other'] || CATEGORY_META.other
  const icon = nodeData.icon || cat.icon
  const color = nodeData.color || cat.color
  const status = nodeData.device?.status || 'unknown'
  const subtitle = [nodeData.device?.manufacturer, nodeData.device?.model].filter(Boolean).join(' ') || cat.label

  return (
    <div className={`telecom-node-card ${selected ? 'selected' : ''}`} style={{ borderTopColor: color }}>
      <Handle type="target" position={Position.Top} id="top" className="telecom-handle" />
      <Handle type="target" position={Position.Left} id="left" className="telecom-handle" />
      <Handle type="source" position={Position.Right} id="right" className="telecom-handle" />
      <Handle type="source" position={Position.Bottom} id="bottom" className="telecom-handle" />

      <div className="node-header">
        <span className="node-icon-badge" style={{ backgroundColor: `${color}25`, color }}>
          {icon}
        </span>
        <div className="node-title-group">
          <strong className="node-title" title={nodeData.label}>{nodeData.label}</strong>
          <small className="node-subtitle">{subtitle}</small>
        </div>
        <span className={`status-dot status-dot-${status}`} title={`Status: ${status}`} />
      </div>

      {(nodeData.ip || nodeData.vlan || nodeData.deviceId) && (
        <div className="node-footer">
          {nodeData.ip && <span className="node-tag ip">{nodeData.ip}</span>}
          {nodeData.vlan && <span className="node-tag vlan">VLAN {nodeData.vlan}</span>}
          {nodeData.deviceId && <span className="node-tag linked" title="Vinculado ao inventário">🔗</span>}
        </div>
      )}
    </div>
  )
}

const nodeTypes = {
  telecomNode: TelecomNodeComponent,
}

const buildNode = (diagramNode: DiagramNode, devices: Device[]): Node => {
  const device = devices.find(d => d.id === diagramNode.deviceId)
  let parsedStyle: Record<string, string> = {}
  try {
    if (diagramNode.styleJson && diagramNode.styleJson !== '{}') {
      parsedStyle = JSON.parse(diagramNode.styleJson)
    }
  } catch {
    parsedStyle = {}
  }

  const category = parsedStyle.category || device?.categoryId || 'other'
  const cat = CATEGORY_META[category] || CATEGORY_META.other

  return {
    id: diagramNode.id,
    type: 'telecomNode',
    position: { x: diagramNode.x, y: diagramNode.y },
    data: {
      label: diagramNode.label || device?.name || 'Equipamento',
      deviceId: diagramNode.deviceId,
      device,
      category,
      icon: parsedStyle.icon || cat.icon,
      color: parsedStyle.color || cat.color,
      ip: parsedStyle.ip || device?.hostname || '',
      mac: parsedStyle.mac || '',
      vlan: parsedStyle.vlan || device?.vlan || '',
      notes: parsedStyle.notes || device?.notes || '',
    } as unknown as Record<string, unknown>,
    style: {
      width: diagramNode.width || 210,
      height: diagramNode.height || 86,
    },
  }
}

const buildEdge = (value: DiagramEdge): Edge => {
  let strokeColor = value.color || '#3b82f6'
  let strokeDash = value.lineStyle === 'dashed' ? '6 4' : value.lineStyle === 'dotted' ? '2 3' : undefined

  if (!value.color) {
    if (value.type === 'Fibra Óptica') strokeColor = '#f59e0b'
    else if (value.type === 'Link WAN') { strokeColor = '#ef4444'; strokeDash = '6 4' }
    else if (value.type === 'Wireless') { strokeColor = '#8b5cf6'; strokeDash = '6 4' }
    else if (value.type === 'Telefonia') strokeColor = '#06b6d4'
  }

  const labelParts = [value.name, value.speed, value.vlan ? `VLAN ${value.vlan}` : ''].filter(Boolean)
  const edgeLabel = labelParts.join(' · ') || value.type || 'Ethernet'

  return {
    id: value.id,
    source: value.sourceNodeId,
    target: value.targetNodeId,
    label: edgeLabel,
    data: { record: value },
    animated: value.type === 'Link WAN' || value.type === 'Fibra Óptica',
    style: {
      stroke: strokeColor,
      strokeWidth: 2.5,
      strokeDasharray: strokeDash,
    },
  }
}

function Editor() {
  const [projects, setProjects] = useState<Project[]>([])
  const [project, setProject] = useState('')
  const [diagrams, setDiagrams] = useState<Diagram[]>([])
  const [diagram, setDiagram] = useState<Diagram>()
  const [devices, setDevices] = useState<Device[]>([])
  const [scanHosts, setScanHosts] = useState<ScanHost[]>([])
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [message, setMessage] = useState('')
  const [paletteTab, setPaletteTab] = useState<'devices' | 'templates' | 'scans'>('devices')
  const [paletteSearch, setPaletteSearch] = useState('')

  // Modais de Edição
  const [editingNode, setEditingNode] = useState<{ node: Node; label: string; deviceId: string; category: string; ip: string; vlan: string; notes: string }>()
  const [editingEdge, setEditingEdge] = useState<{ edge: Edge; record: DiagramEdge }>()
  const [showBOM, setShowBOM] = useState(false)
  const [showCustomNodeModal, setShowCustomNodeModal] = useState(false)
  const [customNodeForm, setCustomNodeForm] = useState({ label: '', category: 'router', ip: '', vlan: '', notes: '' })

  const canvas = useRef<HTMLDivElement>(null)

  // Carregar Projetos
  const loadProjects = async () => {
    try {
      const data = await api<{ items: Project[] }>('/projects')
      setProjects(items(data))
    } catch (err) {
      setMessage(`Erro ao carregar projetos: ${(err as Error).message}`)
    }
  }

  // Carregar Projeto Selecionado
  const loadProject = async (id: string) => {
    setProject(id)
    setDiagram(undefined)
    setNodes([])
    setEdges([])
    if (!id) {
      setDiagrams([])
      setDevices([])
      setScanHosts([])
      return
    }

    try {
      const [d, dv, sc] = await Promise.all([
        api<{ items: Diagram[] }>(`/projects/${id}/diagrams`),
        api<{ items: Device[] }>(`/devices?project_id=${id}`),
        api<{ items: Scan[] }>(`/scans?project_id=${id}`).catch(() => ({ items: [] as Scan[] })),
      ])

      const projectDiagrams = items(d)
      const projectDevices = items(dv)
      setDiagrams(projectDiagrams)
      setDevices(projectDevices)

      // Carregar hosts dos scans do projeto para sugerir na topologia
      const scansList = items(sc)
      if (scansList.length > 0) {
        const lastScan = scansList[0]
        const hData = await api<{ items: ScanHost[] }>(`/scans/${lastScan.id}/hosts`).catch(() => ({ items: [] as ScanHost[] }))
        setScanHosts(items(hData).filter(h => h.status === 'online'))
      } else {
        setScanHosts([])
      }

      // Se houver diagramas, carrega o primeiro automaticamente
      if (projectDiagrams.length > 0) {
        await loadGraphWithDevices(projectDiagrams[0].id, projectDevices)
      }
    } catch (err) {
      setMessage(`Erro ao carregar dados do projeto: ${(err as Error).message}`)
    }
  }

  // Carregar Grafo com lista de dispositivos atualizada
  const loadGraphWithDevices = async (diagramId: string, currentDevices: Device[]) => {
    if (!diagramId) {
      setDiagram(undefined)
      setNodes([])
      setEdges([])
      return
    }
    try {
      const graph = await api<Graph>(`/diagrams/${diagramId}`)
      setDiagram(graph.diagram)
      setNodes(graph.nodes.map(n => buildNode(n, currentDevices)))
      setEdges(graph.edges.map(buildEdge))
    } catch (err) {
      setMessage(`Erro ao abrir diagrama: ${(err as Error).message}`)
    }
  }

  const loadGraph = async (id: string) => {
    await loadGraphWithDevices(id, devices)
  }

  useEffect(() => {
    void loadProjects()
  }, [])

  // Criar Diagrama
  const createDiagram = async () => {
    if (!project) return
    const name = prompt('Nome do novo diagrama:', 'Topologia Geral de Rede')
    if (!name) return
    try {
      const saved = await api<Diagram>(`/projects/${project}/diagrams`, json('POST', { name, description: 'Topologia estruturada' }))
      setDiagrams(current => [...current, saved])
      await loadGraphWithDevices(saved.id, devices)
      setMessage(`Diagrama "${name}" criado.`)
    } catch (err) {
      setMessage(`Erro ao criar diagrama: ${(err as Error).message}`)
    }
  }

  // Adicionar Equipamento do Inventário ao Diagrama
  const addDeviceToDiagram = async (device: Device) => {
    if (!diagram) {
      setMessage('Selecione ou crie um diagrama antes de adicionar equipamentos.')
      return
    }
    try {
      const x = 60 + (nodes.length % 4) * 240
      const y = 60 + Math.floor(nodes.length / 4) * 140
      const catMeta = CATEGORY_META[device.categoryId] || CATEGORY_META.other

      const styleJson = JSON.stringify({
        category: device.categoryId,
        icon: catMeta.icon,
        color: catMeta.color,
        ip: device.hostname || '',
        vlan: device.vlan || '',
        notes: device.notes || '',
      })

      const saved = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes`, json('POST', {
        deviceId: device.id,
        label: device.name,
        x,
        y,
        width: 210,
        height: 86,
        styleJson,
      }))

      setNodes(current => [...current, buildNode(saved, devices)])
      setMessage(`"${device.name}" adicionado à topologia.`)
    } catch (err) {
      setMessage(`Erro ao adicionar nó: ${(err as Error).message}`)
    }
  }

  // Adicionar TODOS os Equipamentos do Projeto de uma vez
  const addAllDevicesToDiagram = async () => {
    if (!diagram) return
    const existingDeviceIds = new Set(nodes.map(n => (n.data as unknown as NodeData).deviceId).filter(Boolean))
    const toAdd = devices.filter(d => !existingDeviceIds.has(d.id))

    if (toAdd.length === 0) {
      setMessage('Todos os equipamentos cadastrados já estão no diagrama.')
      return
    }

    try {
      let currentLen = nodes.length
      const createdNodes: DiagramNode[] = []

      for (const device of toAdd) {
        const x = 60 + (currentLen % 4) * 240
        const y = 60 + Math.floor(currentLen / 4) * 140
        currentLen++

        const catMeta = CATEGORY_META[device.categoryId] || CATEGORY_META.other
        const styleJson = JSON.stringify({
          category: device.categoryId,
          icon: catMeta.icon,
          color: catMeta.color,
          ip: device.hostname || '',
          vlan: device.vlan || '',
        })

        const saved = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes`, json('POST', {
          deviceId: device.id,
          label: device.name,
          x,
          y,
          width: 210,
          height: 86,
          styleJson,
        }))
        createdNodes.push(saved)
      }

      setNodes(current => [...current, ...createdNodes.map(n => buildNode(n, devices))])
      setMessage(`${toAdd.length} equipamentos adicionados ao diagrama.`)
    } catch (err) {
      setMessage(`Erro ao adicionar equipamentos em lote: ${(err as Error).message}`)
    }
  }

  // Adicionar Elemento de Template / Avulso
  const addTemplateNode = async (template: typeof TEMPLATES[0]) => {
    if (!diagram) {
      setMessage('Selecione ou crie um diagrama primeiro.')
      return
    }
    try {
      const x = 60 + (nodes.length % 4) * 240
      const y = 60 + Math.floor(nodes.length / 4) * 140

      const styleJson = JSON.stringify({
        category: template.category,
        icon: template.icon,
        color: template.color,
      })

      const saved = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes`, json('POST', {
        deviceId: '',
        label: template.label,
        x,
        y,
        width: 210,
        height: 86,
        styleJson,
      }))

      setNodes(current => [...current, buildNode(saved, devices)])
      setMessage(`Elemento "${template.label}" adicionado.`)
    } catch (err) {
      setMessage(`Erro ao adicionar elemento: ${(err as Error).message}`)
    }
  }

  // Adicionar a partir de Host Descoberto (Scan)
  const addScanHostToDiagram = async (host: ScanHost) => {
    if (!diagram || !project) return
    const defaultName = host.hostname || `${host.deviceType || 'Dispositivo'} ${host.ip}`
    const name = prompt('Nome para o novo equipamento:', defaultName)
    if (!name) return

    try {
      // 1. Cria o dispositivo no inventário do projeto
      const createdDevice = await api<Device>('/devices', json('POST', {
        projectId: project,
        name,
        categoryId: host.categoryId || 'other',
        manufacturer: host.manufacturer,
        hostname: host.hostname,
        status: 'online',
      }))

      // 2. Adiciona endereço IP
      await api(`/devices/${createdDevice.id}/addresses`, json('POST', {
        type: 'ipv4',
        address: host.ip,
        primary: true,
      }))

      if (host.mac) {
        await api(`/devices/${createdDevice.id}/addresses`, json('POST', {
          type: 'mac',
          address: host.mac,
          primary: true,
        }))
      }

      // 3. Atualiza estado local de devices
      const updatedDevices = [...devices, createdDevice]
      setDevices(updatedDevices)

      // 4. Adiciona nó no diagrama
      const x = 60 + (nodes.length % 4) * 240
      const y = 60 + Math.floor(nodes.length / 4) * 140
      const catMeta = CATEGORY_META[createdDevice.categoryId] || CATEGORY_META.other

      const styleJson = JSON.stringify({
        category: createdDevice.categoryId,
        icon: catMeta.icon,
        color: catMeta.color,
        ip: host.ip,
        mac: host.mac,
      })

      const savedNode = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes`, json('POST', {
        deviceId: createdDevice.id,
        label: createdDevice.name,
        x,
        y,
        width: 210,
        height: 86,
        styleJson,
      }))

      setNodes(current => [...current, buildNode(savedNode, updatedDevices)])
      setMessage(`Host "${name}" (${host.ip}) cadastrado no inventário e adicionado à topologia.`)
    } catch (err) {
      setMessage(`Erro ao importar host para topologia: ${(err as Error).message}`)
    }
  }

  // Adicionar Nó Customizado via Modal
  const submitCustomNode = async () => {
    if (!diagram || !customNodeForm.label.trim()) return
    const cat = CATEGORY_META[customNodeForm.category] || CATEGORY_META.other
    try {
      const x = 60 + (nodes.length % 4) * 240
      const y = 60 + Math.floor(nodes.length / 4) * 140

      const styleJson = JSON.stringify({
        category: customNodeForm.category,
        icon: cat.icon,
        color: cat.color,
        ip: customNodeForm.ip,
        vlan: customNodeForm.vlan,
        notes: customNodeForm.notes,
      })

      const saved = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes`, json('POST', {
        deviceId: '',
        label: customNodeForm.label.trim(),
        x,
        y,
        width: 210,
        height: 86,
        styleJson,
      }))

      setNodes(current => [...current, buildNode(saved, devices)])
      setShowCustomNodeModal(false)
      setCustomNodeForm({ label: '', category: 'router', ip: '', vlan: '', notes: '' })
      setMessage(`Nó "${saved.label}" criado com sucesso.`)
    } catch (err) {
      setMessage(`Erro ao criar nó customizado: ${(err as Error).message}`)
    }
  }

  // React Flow Handlers
  const changeNodes = useCallback((changes: NodeChange[]) => {
    setNodes(current => applyNodeChanges(changes, current))
  }, [])

  const changeEdges = useCallback((changes: unknown[]) => {
    setEdges(current => applyEdgeChanges(changes as any, current))
  }, [])

  const persistNodePosition = async (_: unknown, node: Node) => {
    if (!diagram) return
    const nodeData = node.data as unknown as NodeData
    const cat = CATEGORY_META[nodeData.category || 'other'] || CATEGORY_META.other

    const styleJson = JSON.stringify({
      category: nodeData.category,
      icon: nodeData.icon || cat.icon,
      color: nodeData.color || cat.color,
      ip: nodeData.ip,
      mac: nodeData.mac,
      vlan: nodeData.vlan,
      notes: nodeData.notes,
    })

    try {
      await api(`/diagrams/${diagram.id}/nodes/${node.id}`, json('PUT', {
        deviceId: nodeData.deviceId || '',
        label: nodeData.label || 'Equipamento',
        x: node.position.x,
        y: node.position.y,
        width: Number(node.measured?.width || node.style?.width || 210),
        height: Number(node.measured?.height || node.style?.height || 86),
        styleJson,
      }))
    } catch (err) {
      console.error('Falha ao salvar posição do nó:', err)
    }
  }

  // Conectar 2 Nós
  const connect = useCallback(async (connection: Connection) => {
    if (!diagram || !connection.source || !connection.target) return
    try {
      const saved = await api<DiagramEdge>(`/diagrams/${diagram.id}/edges`, json('POST', {
        sourceNodeId: connection.source,
        targetNodeId: connection.target,
        name: 'Ethernet',
        type: 'Ethernet',
        speed: '1 Gbps',
        color: '#3b82f6',
        lineStyle: 'solid',
      }))
      setEdges(current => addEdge(buildEdge(saved), current))
      setMessage('Ligação criada com sucesso.')
    } catch (err) {
      setMessage(`Erro ao criar conexão: ${(err as Error).message}`)
    }
  }, [diagram])

  // Remover Seleção (Tratando Cascade com Segurança)
  const removeSelected = async () => {
    if (!diagram) return
    const selectedEdges = edges.filter(e => e.selected)
    const selectedNodes = nodes.filter(n => n.selected)

    if (selectedEdges.length === 0 && selectedNodes.length === 0) {
      setMessage('Nenhum elemento selecionado para exclusão.')
      return
    }

    try {
      // 1. Exclui primeiro os nós (o SQLite já cascateia arestas conectadas)
      for (const node of selectedNodes) {
        await api(`/diagrams/${diagram.id}/nodes/${node.id}`, { method: 'DELETE' }).catch(() => {})
      }

      // 2. Exclui arestas selecionadas que não estejam ligadas aos nós já excluídos
      const deletedNodeIds = new Set(selectedNodes.map(n => n.id))
      for (const edge of selectedEdges) {
        if (!deletedNodeIds.has(edge.source) && !deletedNodeIds.has(edge.target)) {
          await api(`/diagrams/${diagram.id}/edges/${edge.id}`, { method: 'DELETE' }).catch(() => {})
        }
      }

      // 3. Atualiza estado da UI
      setEdges(current => current.filter(e => !e.selected && !deletedNodeIds.has(e.source) && !deletedNodeIds.has(e.target)))
      setNodes(current => current.filter(n => !n.selected))
      setMessage('Elementos selecionados removidos.')
    } catch (err) {
      setMessage(`Erro na exclusão: ${(err as Error).message}`)
    }
  }

  // Auto Layout Hierárquico Inteligente
  const autoLayout = async (type: 'hierarchical' | 'grid' = 'hierarchical') => {
    if (!diagram || nodes.length === 0) return

    let arranged: Node[] = []

    if (type === 'grid') {
      const cols = Math.max(3, Math.ceil(Math.sqrt(nodes.length * 1.5)))
      arranged = nodes.map((node, i) => ({
        ...node,
        position: {
          x: 60 + (i % cols) * 250,
          y: 60 + Math.floor(i / cols) * 150,
        },
      }))
    } else {
      // Organização Hierárquica por Categoria de Rede
      const layers: { [k: string]: Node[] } = {
        wan: [],
        router: [],
        switch: [],
        endpoints: [],
      }

      nodes.forEach(node => {
        const cat = (node.data as unknown as NodeData).category || 'other'
        if (cat === 'gateway') layers.wan.push(node)
        else if (['router', 'firewall', 'olt'].includes(cat)) layers.router.push(node)
        else if (['switch', 'patch-panel'].includes(cat)) layers.switch.push(node)
        else layers.endpoints.push(node)
      })

      const allLayers = [layers.wan, layers.router, layers.switch, layers.endpoints].filter(l => l.length > 0)
      const maxInLayer = Math.max(...allLayers.map(l => l.length), 1)
      const canvasCenter = (maxInLayer * 260) / 2

      let currentY = 60
      arranged = []

      allLayers.forEach(layer => {
        const layerWidth = layer.length * 260
        const startX = Math.max(60, canvasCenter - layerWidth / 2 + 60)

        layer.forEach((node, idx) => {
          arranged.push({
            ...node,
            position: {
              x: startX + idx * 260,
              y: currentY,
            },
          })
        })
        currentY += 160
      })
    }

    setNodes(arranged)

    // Persiste novas posições
    try {
      await Promise.all(arranged.map(node => {
        const nodeData = node.data as unknown as NodeData
        const cat = CATEGORY_META[nodeData.category || 'other'] || CATEGORY_META.other
        const styleJson = JSON.stringify({
          category: nodeData.category,
          icon: nodeData.icon || cat.icon,
          color: nodeData.color || cat.color,
          ip: nodeData.ip,
          vlan: nodeData.vlan,
          notes: nodeData.notes,
        })
        return api(`/diagrams/${diagram.id}/nodes/${node.id}`, json('PUT', {
          deviceId: nodeData.deviceId || '',
          label: nodeData.label || 'Equipamento',
          x: node.position.x,
          y: node.position.y,
          width: 210,
          height: 86,
          styleJson,
        }))
      }))
      setMessage('Layout organizado com sucesso.')
    } catch (err) {
      console.error('Erro ao salvar layout:', err)
    }
  }

  // Exportar PNG / PDF
  const exportRaster = async (pdf = false) => {
    const viewport = canvas.current?.querySelector('.react-flow__viewport') as HTMLElement | null
    if (!viewport || nodes.length === 0) {
      setMessage('O diagrama está vazio.')
      return
    }

    try {
      const bounds = getNodesBounds(nodes)
      const width = Math.max(1400, bounds.width + 200)
      const height = Math.max(900, bounds.height + 200)
      const view = getViewportForBounds(bounds, width, height, 0.3, 2, 0.15)

      const isLight = document.documentElement.dataset.theme === 'light'
      const bgColor = isLight ? '#f4f7fb' : '#08111f'

      const data = await toPng(viewport, {
        backgroundColor: bgColor,
        pixelRatio: 2,
        width,
        height,
        style: {
          width: `${width}px`,
          height: `${height}px`,
          transform: `translate(${view.x}px, ${view.y}px) scale(${view.zoom})`,
        },
      })

      if (pdf) {
        const out = new jsPDF({ orientation: 'landscape', unit: 'mm', format: 'a4' })
        const p = out.getImageProperties(data)
        const pageWidth = 297 - 20
        const pageHeight = 210 - 20

        const ratio = Math.min(pageWidth / p.width, pageHeight / p.height)
        const renderW = p.width * ratio
        const renderH = p.height * ratio
        const posX = 10 + (pageWidth - renderW) / 2
        const posY = 10 + (pageHeight - renderH) / 2

        out.addImage(data, 'PNG', posX, posY, renderW, renderH)
        out.save(`${diagram?.name || 'topologia'}.pdf`)
        setMessage('PDF exportado com sucesso.')
      } else {
        const link = document.createElement('a')
        link.download = `${diagram?.name || 'topologia'}.png`
        link.href = data
        link.click()
        setMessage('Imagem PNG exportada.')
      }
    } catch (err) {
      setMessage(`Erro ao exportar: ${(err as Error).message}`)
    }
  }

  // Double click no Nó para Edição
  const onNodeDoubleClick = (_: unknown, node: Node) => {
    const data = node.data as unknown as NodeData
    setEditingNode({
      node,
      label: data.label || '',
      deviceId: data.deviceId || '',
      category: data.category || 'other',
      ip: data.ip || '',
      vlan: data.vlan || '',
      notes: data.notes || '',
    })
  }

  // Salvar Edição do Nó
  const saveNodeEdit = async () => {
    if (!diagram || !editingNode) return
    const cat = CATEGORY_META[editingNode.category] || CATEGORY_META.other
    const linkedDevice = devices.find(d => d.id === editingNode.deviceId)

    const styleJson = JSON.stringify({
      category: editingNode.category,
      icon: cat.icon,
      color: cat.color,
      ip: editingNode.ip,
      vlan: editingNode.vlan,
      notes: editingNode.notes,
    })

    try {
      const updated = await api<DiagramNode>(`/diagrams/${diagram.id}/nodes/${editingNode.node.id}`, json('PUT', {
        deviceId: editingNode.deviceId || '',
        label: editingNode.label.trim() || 'Equipamento',
        x: editingNode.node.position.x,
        y: editingNode.node.position.y,
        width: 210,
        height: 86,
        styleJson,
      }))

      setNodes(current => current.map(n => n.id === updated.id ? buildNode(updated, devices) : n))
      setEditingNode(undefined)
      setMessage(`Elemento "${updated.label}" atualizado.`)
    } catch (err) {
      setMessage(`Erro ao atualizar nó: ${(err as Error).message}`)
    }
  }

  // Double click na Ligação para Edição
  const onEdgeDoubleClick = (_: unknown, edge: Edge) => {
    const record = edge.data?.record as DiagramEdge
    if (!record) return
    setEditingEdge({
      edge,
      record: { ...record },
    })
  }

  // Salvar Edição da Ligação
  const saveEdgeEdit = async () => {
    if (!diagram || !editingEdge) return
    try {
      const saved = await api<DiagramEdge>(`/diagrams/${diagram.id}/edges/${editingEdge.record.id}`, json('PUT', editingEdge.record))
      setEdges(current => current.map(e => e.id === saved.id ? buildEdge(saved) : e))
      setEditingEdge(undefined)
      setMessage('Ligação atualizada com sucesso.')
    } catch (err) {
      setMessage(`Erro ao atualizar ligação: ${(err as Error).message}`)
    }
  }

  // Filtros da Paleta
  const filteredDevices = useMemo(() => {
    const existingDeviceIds = new Set(nodes.map(n => (n.data as unknown as NodeData).deviceId).filter(Boolean))
    return devices.filter(d => {
      const matchesSearch = !paletteSearch || d.name.toLowerCase().includes(paletteSearch.toLowerCase()) || d.hostname.toLowerCase().includes(paletteSearch.toLowerCase()) || d.manufacturer.toLowerCase().includes(paletteSearch.toLowerCase())
      return matchesSearch
    }).map(d => ({
      ...d,
      inDiagram: existingDeviceIds.has(d.id),
    }))
  }, [devices, nodes, paletteSearch])

  const filteredTemplates = useMemo(() => {
    return TEMPLATES.filter(t => !paletteSearch || t.label.toLowerCase().includes(paletteSearch.toLowerCase()) || t.category.toLowerCase().includes(paletteSearch.toLowerCase()))
  }, [paletteSearch])

  const filteredScans = useMemo(() => {
    return scanHosts.filter(h => !paletteSearch || h.ip.includes(paletteSearch) || h.hostname.toLowerCase().includes(paletteSearch.toLowerCase()) || h.manufacturer.toLowerCase().includes(paletteSearch.toLowerCase()))
  }, [scanHosts, paletteSearch])

  // Cálculo de Resumo / Orçamento (BOM)
  const bomSummary = useMemo(() => {
    const catCounts: Record<string, number> = {}
    const speedCounts: Record<string, number> = {}
    const typeCounts: Record<string, number> = {}

    nodes.forEach(node => {
      const data = node.data as unknown as NodeData
      const cat = data.category || 'other'
      catCounts[cat] = (catCounts[cat] || 0) + 1
    })

    edges.forEach(edge => {
      const record = edge.data?.record as DiagramEdge
      const type = record?.type || 'Ethernet'
      const speed = record?.speed || '1 Gbps'
      typeCounts[type] = (typeCounts[type] || 0) + 1
      speedCounts[speed] = (speedCounts[speed] || 0) + 1
    })

    return {
      totalNodes: nodes.length,
      totalEdges: edges.length,
      catCounts,
      typeCounts,
      speedCounts,
    }
  }, [nodes, edges])

  return (
    <>
      <PageHeader
        eyebrow="REDE & TELECOM"
        title="Topologia e Diagramas de Rede"
        description="Modelagem visual completa com infraestrutura, switches, roteadores, CFTV, links e orçamentos."
        actions={
          <>
            <button disabled={!diagram} onClick={() => void autoLayout('hierarchical')} title="Organizar em camadas lógicas (WAN > Roteadores > Switches > Pontas)">
              ☷ Organizar Hierarquia
            </button>
            <button disabled={!diagram} onClick={() => void autoLayout('grid')} title="Organizar em grade compacta">
              ▦ Grade
            </button>
            <button disabled={!diagram} onClick={() => void removeSelected()} className="danger" title="Remover nós ou ligações selecionados">
              × Remover Seleção
            </button>
            <button disabled={!diagram} onClick={() => setShowBOM(true)} title="Visualizar lista de materiais e resumo da topologia">
              📋 Resumo / Orçamento
            </button>
            <button disabled={!diagram} onClick={() => void exportRaster(false)}>
              PNG
            </button>
            <button disabled={!diagram} onClick={() => void exportRaster(true)}>
              PDF
            </button>
            {diagram && (
              <a className="button" href={`/api/v1/diagrams/${diagram.id}/export.svg`} download={`${diagram.name}.svg`}>
                SVG
              </a>
            )}
          </>
        }
      />

      <Notice message={message} onClose={() => setMessage('')} />

      <div className="topology-toolbar">
        <select value={project} onChange={e => void loadProject(e.target.value)}>
          <option value="">Selecione um projeto</option>
          {projects.map(value => (
            <option key={value.id} value={value.id}>{value.name}</option>
          ))}
        </select>

        <select value={diagram?.id || ''} onChange={e => void loadGraph(e.target.value)} disabled={!project || diagrams.length === 0}>
          <option value="">{diagrams.length === 0 ? 'Nenhum diagrama no projeto' : 'Selecione o diagrama'}</option>
          {diagrams.map(value => (
            <option key={value.id} value={value.id}>{value.name}</option>
          ))}
        </select>

        <button disabled={!project} onClick={() => void createDiagram()} className="primary">
          + Novo Diagrama
        </button>

        {diagram && (
          <button onClick={() => setShowCustomNodeModal(true)}>
            + Elemento Personalizado
          </button>
        )}
      </div>

      <div className="topology-layout">
        {/* Painel Lateral de Entidades & Paleta */}
        <Panel className="palette-panel" title="Paleta de Entidades">
          <div className="palette-tabs">
            <button className={paletteTab === 'devices' ? 'active' : ''} onClick={() => setPaletteTab('devices')}>
              Equipamentos ({devices.length})
            </button>
            <button className={paletteTab === 'templates' ? 'active' : ''} onClick={() => setPaletteTab('templates')}>
              Telecom / Infra
            </button>
            {scanHosts.length > 0 && (
              <button className={paletteTab === 'scans' ? 'active' : ''} onClick={() => setPaletteTab('scans')}>
                Scan ({scanHosts.length})
              </button>
            )}
          </div>

          <div className="palette-search">
            <input
              value={paletteSearch}
              onChange={e => setPaletteSearch(e.target.value)}
              placeholder="Filtrar entidades..."
            />
          </div>

          {paletteTab === 'devices' && (
            <div className="palette-content">
              {devices.length > 0 && (
                <button className="add-all-btn" onClick={() => void addAllDevicesToDiagram()} disabled={!diagram}>
                  + Adicionar Todos ao Diagrama
                </button>
              )}
              <div className="device-palette">
                {filteredDevices.length === 0 ? (
                  <small className="palette-empty">Nenhum equipamento encontrado.</small>
                ) : (
                  filteredDevices.map(device => {
                    const cat = CATEGORY_META[device.categoryId] || CATEGORY_META.other
                    return (
                      <button
                        key={device.id}
                        disabled={!diagram || device.inDiagram}
                        onClick={() => void addDeviceToDiagram(device)}
                        className={`palette-item ${device.inDiagram ? 'in-diagram' : ''}`}
                      >
                        <div className="palette-item-header">
                          <span className="palette-icon" style={{ color: cat.color }}>{cat.icon}</span>
                          <strong>{device.name}</strong>
                          {device.inDiagram && <span className="in-badge">No diagrama</span>}
                        </div>
                        <span>{[device.manufacturer, device.model].filter(Boolean).join(' ') || cat.label}</span>
                        {device.hostname && <small className="palette-ip">{device.hostname}</small>}
                      </button>
                    )
                  })
                )}
              </div>
            </div>
          )}

          {paletteTab === 'templates' && (
            <div className="palette-content">
              <div className="template-grid">
                {filteredTemplates.map(t => (
                  <button
                    key={t.key}
                    disabled={!diagram}
                    onClick={() => void addTemplateNode(t)}
                    className="template-card"
                  >
                    <span className="template-icon" style={{ color: t.color }}>{t.icon}</span>
                    <b>{t.label}</b>
                  </button>
                ))}
              </div>
            </div>
          )}

          {paletteTab === 'scans' && (
            <div className="palette-content">
              <p className="palette-hint">Hosts online do último scan no projeto. Clique para cadastrar e inserir no diagrama:</p>
              <div className="device-palette">
                {filteredScans.map((host, idx) => (
                  <button
                    key={idx}
                    disabled={!diagram}
                    onClick={() => void addScanHostToDiagram(host)}
                    className="palette-item scan-item"
                  >
                    <div className="palette-item-header">
                      <span className="palette-icon">📡</span>
                      <strong>{host.hostname || host.ip}</strong>
                    </div>
                    <span>{host.manufacturer || host.deviceType || 'Dispositivo'} · {host.ip}</span>
                    {host.mac && <small className="palette-ip">MAC: {host.mac}</small>}
                  </button>
                ))}
              </div>
            </div>
          )}
        </Panel>

        {/* Canvas Visual do React Flow */}
        <Panel className="topology-panel">
          {!diagram ? (
            <Empty>
              Selecione um projeto e diagrama acima, ou crie um novo diagrama para começar.
            </Empty>
          ) : (
            <div className="flow-canvas-container" ref={canvas}>
              <div className="canvas-header-info">
                <span><b>{diagram.name}</b> · {nodes.length} nós · {edges.length} conexões</span>
                <small>Dica: Arraste os pontos para conectar nós. Dê duplo clique em nós ou arestas para editar detalhes.</small>
              </div>
              <ReactFlow
                nodes={nodes}
                edges={edges}
                nodeTypes={nodeTypes}
                onNodesChange={changeNodes}
                onEdgesChange={changeEdges}
                onNodeDragStop={(event, node) => void persistNodePosition(event, node)}
                onNodeDoubleClick={onNodeDoubleClick}
                onEdgeDoubleClick={onEdgeDoubleClick}
                onConnect={connect}
                fitView
                snapToGrid
                snapGrid={[20, 20]}
                deleteKeyCode={['Backspace', 'Delete']}
              >
                <Background gap={20} size={1} />
                <MiniMap
                  nodeColor={n => {
                    const data = n.data as unknown as NodeData
                    const cat = CATEGORY_META[data.category || 'other'] || CATEGORY_META.other
                    return cat.color || '#3b82f6'
                  }}
                  maskColor="rgba(8, 17, 31, 0.6)"
                />
                <Controls />
              </ReactFlow>
            </div>
          )}
        </Panel>
      </div>

      {/* Modal: Editar Nó */}
      {editingNode && (
        <Modal title="Editar Elemento da Topologia" onClose={() => setEditingNode(undefined)}>
          <div className="form-grid">
            <Field label="Nome / Rótulo">
              <input
                value={editingNode.label}
                onChange={e => setEditingNode({ ...editingNode, label: e.target.value })}
                placeholder="Ex: Roteador Borda, Switch Core 24p"
              />
            </Field>

            <Field label="Categoria de Rede">
              <select
                value={editingNode.category}
                onChange={e => setEditingNode({ ...editingNode, category: e.target.value })}
              >
                {Object.entries(CATEGORY_META).map(([key, meta]) => (
                  <option key={key} value={key}>{meta.icon} {meta.label}</option>
                ))}
              </select>
            </Field>

            <Field label="Vincular a Equipamento do Inventário">
              <select
                value={editingNode.deviceId}
                onChange={e => {
                  const devId = e.target.value
                  const dev = devices.find(d => d.id === devId)
                  setEditingNode({
                    ...editingNode,
                    deviceId: devId,
                    label: dev ? dev.name : editingNode.label,
                    category: dev ? dev.categoryId : editingNode.category,
                    ip: dev?.hostname || editingNode.ip,
                    vlan: dev?.vlan || editingNode.vlan,
                  })
                }}
              >
                <option value="">Sem vínculo (Elemento lógico / template)</option>
                {devices.map(d => (
                  <option key={d.id} value={d.id}>{d.name} ({d.manufacturer || d.categoryId})</option>
                ))}
              </select>
            </Field>

            <Field label="Endereço IP / Hostname">
              <input
                value={editingNode.ip}
                onChange={e => setEditingNode({ ...editingNode, ip: e.target.value })}
                placeholder="Ex: 192.168.1.1 ou srv-db.local"
              />
            </Field>

            <Field label="VLAN">
              <input
                value={editingNode.vlan}
                onChange={e => setEditingNode({ ...editingNode, vlan: e.target.value })}
                placeholder="Ex: 10, 100, Trunk"
              />
            </Field>

            <Field label="Observações">
              <textarea
                value={editingNode.notes}
                onChange={e => setEditingNode({ ...editingNode, notes: e.target.value })}
                placeholder="Portas, localização no rack, IP secundário..."
              />
            </Field>

            <div className="form-actions">
              <button type="button" onClick={() => setEditingNode(undefined)}>Cancelar</button>
              <button type="button" className="primary" onClick={() => void saveNodeEdit()}>Salvar Alterações</button>
            </div>
          </div>
        </Modal>
      )}

      {/* Modal: Editar Ligação (Edge) */}
      {editingEdge && (
        <Modal title="Editar Ligação de Rede" onClose={() => setEditingEdge(undefined)}>
          <div className="form-grid">
            <Field label="Rótulo / Identificação">
              <input
                value={editingEdge.record.name}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, name: e.target.value },
                })}
                placeholder="Ex: Backbone Fibra, Trunk PoE, Link WAN Vivo"
              />
            </Field>

            <Field label="Meio de Transmissão / Tipo">
              <select
                value={editingEdge.record.type}
                onChange={e => {
                  const val = e.target.value
                  let color = editingEdge.record.color
                  let lineStyle = editingEdge.record.lineStyle
                  if (val === 'Fibra Óptica') color = '#f59e0b'
                  else if (val === 'Ethernet') color = '#3b82f6'
                  else if (val === 'Link WAN') { color = '#ef4444'; lineStyle = 'dashed' }
                  else if (val === 'Wireless') { color = '#8b5cf6'; lineStyle = 'dashed' }
                  else if (val === 'Telefonia') color = '#06b6d4'

                  setEditingEdge({
                    ...editingEdge,
                    record: { ...editingEdge.record, type: val, color, lineStyle },
                  })
                }}
              >
                <option value="Ethernet">Ethernet (Cabo de Rede Cat5e / Cat6 / Cat6a)</option>
                <option value="Fibra Óptica">Fibra Óptica (Monomodo / Multimodo)</option>
                <option value="Link WAN">Link WAN / Provedor de Internet</option>
                <option value="Wireless">Wireless / Wi-Fi PTP</option>
                <option value="Telefonia">Telefonia / Cabo Telefônico / VoIP</option>
                <option value="Coaxial">Coaxial / CFTV</option>
              </select>
            </Field>

            <Field label="Velocidade Nominal">
              <select
                value={editingEdge.record.speed}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, speed: e.target.value },
                })}
              >
                <option value="100 Mbps">100 Mbps (Fast Ethernet)</option>
                <option value="300 Mbps">300 Mbps</option>
                <option value="500 Mbps">500 Mbps</option>
                <option value="1 Gbps">1 Gbps (Gigabit Ethernet)</option>
                <option value="2.5 Gbps">2.5 Gbps (Multi-Gigabit)</option>
                <option value="10 Gbps">10 Gbps (10G SFP+ / Fibra)</option>
                <option value="40 Gbps">40 Gbps</option>
                <option value="100 Gbps">100 Gbps</option>
              </select>
            </Field>

            <Field label="VLAN Associada">
              <input
                value={editingEdge.record.vlan}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, vlan: e.target.value },
                })}
                placeholder="Ex: 10, 20, Trunk, 802.1Q"
              />
            </Field>

            <Field label="Porta Origem">
              <input
                value={editingEdge.record.sourceInterface}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, sourceInterface: e.target.value },
                })}
                placeholder="Ex: Port 24, SFP+ 1, eth0"
              />
            </Field>

            <Field label="Porta Destino">
              <input
                value={editingEdge.record.targetInterface}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, targetInterface: e.target.value },
                })}
                placeholder="Ex: WAN1, Port 1, Gi0/1"
              />
            </Field>

            <Field label="Estilo da Linha">
              <select
                value={editingEdge.record.lineStyle}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, lineStyle: e.target.value },
                })}
              >
                <option value="solid">Contínua (Sólida)</option>
                <option value="dashed">Tracejada (Dashed)</option>
                <option value="dotted">Pontilhada (Dotted)</option>
              </select>
            </Field>

            <Field label="Cor">
              <input
                type="color"
                value={editingEdge.record.color || '#3b82f6'}
                onChange={e => setEditingEdge({
                  ...editingEdge,
                  record: { ...editingEdge.record, color: e.target.value },
                })}
              />
            </Field>

            <div className="form-actions">
              <button type="button" onClick={() => setEditingEdge(undefined)}>Cancelar</button>
              <button type="button" className="primary" onClick={() => void saveEdgeEdit()}>Salvar Ligação</button>
            </div>
          </div>
        </Modal>
      )}

      {/* Modal: Adicionar Elemento Customizado */}
      {showCustomNodeModal && (
        <Modal title="Novo Elemento de Rede / Telecom" onClose={() => setShowCustomNodeModal(false)}>
          <div className="form-grid">
            <Field label="Nome do Elemento">
              <input
                required
                value={customNodeForm.label}
                onChange={e => setCustomNodeForm({ ...customNodeForm, label: e.target.value })}
                placeholder="Ex: Roteador BGP, Switch Andar 2, Câmera Estacionamento"
              />
            </Field>

            <Field label="Tipo / Categoria">
              <select
                value={customNodeForm.category}
                onChange={e => setCustomNodeForm({ ...customNodeForm, category: e.target.value })}
              >
                {Object.entries(CATEGORY_META).map(([k, meta]) => (
                  <option key={k} value={k}>{meta.icon} {meta.label}</option>
                ))}
              </select>
            </Field>

            <Field label="Endereço IP">
              <input
                value={customNodeForm.ip}
                onChange={e => setCustomNodeForm({ ...customNodeForm, ip: e.target.value })}
                placeholder="192.168.0.1"
              />
            </Field>

            <Field label="VLAN">
              <input
                value={customNodeForm.vlan}
                onChange={e => setCustomNodeForm({ ...customNodeForm, vlan: e.target.value })}
                placeholder="VLAN 10"
              />
            </Field>

            <Field label="Observações">
              <textarea
                value={customNodeForm.notes}
                onChange={e => setCustomNodeForm({ ...customNodeForm, notes: e.target.value })}
                placeholder="Detalhes adicionais, modelo ou especificações..."
              />
            </Field>

            <div className="form-actions">
              <button type="button" onClick={() => setShowCustomNodeModal(false)}>Cancelar</button>
              <button type="button" className="primary" onClick={() => void submitCustomNode()}>Inserir no Diagrama</button>
            </div>
          </div>
        </Modal>
      )}

      {/* Modal: Resumo de Topologia & Lista de Materiais (BOM) */}
      {showBOM && (
        <Modal title="Resumo da Topologia e Lista de Materiais (BOM)" onClose={() => setShowBOM(false)} wide>
          <div className="bom-container">
            <div className="metric-grid">
              <div className="metric">
                <div className="metric-icon">▣</div>
                <div>
                  <small>Total de Elementos</small>
                  <strong>{bomSummary.totalNodes}</strong>
                </div>
              </div>
              <div className="metric">
                <div className="metric-icon">⇄</div>
                <div>
                  <small>Conexões / Cabos</small>
                  <strong>{bomSummary.totalEdges}</strong>
                </div>
              </div>
              <div className="metric">
                <div className="metric-icon">✓</div>
                <div>
                  <small>Equipamentos no Inventário</small>
                  <strong>{devices.length}</strong>
                </div>
              </div>
            </div>

            <div className="cards two">
              <Panel title="Quantitativo por Categoria">
                <table className="bom-table">
                  <thead>
                    <tr>
                      <th>Categoria</th>
                      <th>Quantidade</th>
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(bomSummary.catCounts).map(([catKey, count]) => {
                      const meta = CATEGORY_META[catKey] || CATEGORY_META.other
                      return (
                        <tr key={catKey}>
                          <td>
                            <span className="bom-cat-icon" style={{ color: meta.color }}>{meta.icon}</span>
                            {meta.label}
                          </td>
                          <td><b>{count} un</b></td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </Panel>

              <Panel title="Conexões por Meio / Velocidade">
                <table className="bom-table">
                  <thead>
                    <tr>
                      <th>Tipo de Enlace</th>
                      <th>Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(bomSummary.typeCounts).map(([type, count]) => (
                      <tr key={type}>
                        <td>{type}</td>
                        <td><b>{count} ligações</b></td>
                      </tr>
                    ))}
                    {Object.entries(bomSummary.speedCounts).map(([speed, count]) => (
                      <tr key={speed}>
                        <td>Velocidade {speed}</td>
                        <td>{count} enlaces</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </Panel>
            </div>

            <div className="form-actions">
              <button type="button" onClick={() => setShowBOM(false)}>Fechar</button>
              <button type="button" className="primary" onClick={() => void exportRaster(true)}>
                Exportar PDF do Diagrama
              </button>
            </div>
          </div>
        </Modal>
      )}
    </>
  )
}

export default function TopologyPage() {
  return (
    <ReactFlowProvider>
      <Editor />
    </ReactFlowProvider>
  )
}
