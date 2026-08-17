import { useEffect, useState } from 'react'
import { api } from '../api'
import { Summary } from '../types'
import { Empty, Page, PageHeader, Panel, Status } from '../components'

type ActionItem={id:string;type:string;title:string;subtitle:string;status:string;dueAt:string;projectId:string}
type Operations={visits:ActionItem[];pending:ActionItem[];drafts:number;returns:number;overdue:number}
const formatDate=(value:string)=>value?new Date(value.includes('T')?value:`${value}T00:00`).toLocaleString('pt-BR',{dateStyle:'short',timeStyle:value.includes('T')?'short':undefined}):'Sem prazo'

export default function DashboardPage({projectId,onNavigate}:{projectId:string;onNavigate:(page:Page,id?:string,entityType?:string)=>void}){
  const[summary,setSummary]=useState<Summary>()
  const[operations,setOperations]=useState<Operations>()
  useEffect(()=>{const query=projectId?`?project_id=${encodeURIComponent(projectId)}`:'';void Promise.all([api<Summary>(`/dashboard${query}`).then(setSummary),api<Operations>(`/dashboard/operations${query}`).then(setOperations)])},[projectId])
  const cards=[['Equipamentos',summary?.devices,'▣'],['Online',summary?.online,'✓'],['Offline',summary?.offline,'×'],['Visitas em rascunho',operations?.drafts,'✎'],['Retornos necessários',operations?.returns,'↩'],['Pendências vencidas',operations?.overdue,'!']]
  return <>
    <PageHeader eyebrow="ROTINA DO TÉCNICO" title="Visão de hoje" description="Atendimentos e pendências que precisam de ação." actions={<><button onClick={()=>onNavigate('scanner')}>Iniciar scanner</button><button className="primary" onClick={()=>onNavigate('visits','new','new')}>+ Nova visita</button></>}/>
    <div className="metric-grid operational-metrics">{cards.map(([label,value,icon])=><article className="metric" key={String(label)}><span className="metric-icon">{icon}</span><div><small>{label}</small><strong>{value??'—'}</strong></div></article>)}</div>
    <div className="cards two operational-grid"><Panel title="Atendimentos de hoje">{!operations?.visits.length?<Empty>Nenhum atendimento previsto para hoje.</Empty>:<div className="action-list">{operations.visits.map(item=><button key={item.id} onClick={()=>onNavigate('visits',item.id,'technical_visit')}><div><b>{item.title}</b><span>{item.subtitle}</span></div><div><Status value={item.status}/><small>{formatDate(item.dueAt)}</small></div></button>)}</div>}</Panel><Panel title="Pendências abertas">{!operations?.pending.length?<Empty>Nenhuma pendência aberta.</Empty>:<div className="action-list">{operations.pending.map(item=><div className="action-row" key={item.id}><div><b>{item.title}</b><span>{item.subtitle}</span></div><small className={item.dueAt&&new Date(`${item.dueAt}T23:59`)<new Date()?'overdue':''}>{formatDate(item.dueAt)}</small></div>)}</div>}</Panel></div>
  </>
}
