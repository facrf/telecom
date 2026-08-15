import { ReactNode, useEffect, useState } from 'react'
import { api, items } from './api'

export type Page = 'dashboard'|'clients'|'visits'|'inventory'|'scanner'|'topology'|'documentation'|'transfer'|'settings'
const navigation: {page:Page;label:string;icon:string}[] = [
  {page:'dashboard',label:'Dashboard',icon:'▦'},{page:'clients',label:'Clientes e projetos',icon:'◉'},
  {page:'visits',label:'Visitas técnicas',icon:'✓'},
  {page:'scanner',label:'Scanner',icon:'⌁'},{page:'inventory',label:'Equipamentos',icon:'▣'},
  {page:'topology',label:'Topologia',icon:'⌘'},{page:'documentation',label:'Documentação',icon:'▤'},
  {page:'transfer',label:'Importar / Backups',icon:'⇄'},{page:'settings',label:'Configurações',icon:'⚙'},
]

export function Shell({page,onPage,theme,onTheme,children}:{page:Page;onPage:(page:Page)=>void;theme:string;onTheme:()=>void;children:ReactNode}) {
  const [query,setQuery]=useState('');const [results,setResults]=useState<{type:string;id:string;title:string;subtitle:string}[]>([])
  useEffect(()=>{if(query.trim().length<2){setResults([]);return};const timer=setTimeout(()=>void api<{items:typeof results}>(`/search?q=${encodeURIComponent(query)}`).then(value=>setResults(items(value))).catch(()=>setResults([])),250);return()=>clearTimeout(timer)},[query])
  return <div className="shell"><aside className="sidebar"><div className="brand"><span className="brand-mark">T</span><div>TELECOM<small>GESTÃO LOCAL</small></div></div><nav>{navigation.map(item=><button key={item.page} className={page===item.page?'nav-active':''} onClick={()=>onPage(item.page)}><span>{item.icon}</span>{item.label}</button>)}</nav><div className="sidebar-footer"><button onClick={onTheme}>{theme==='dark'?'☀ Modo claro':'◐ Modo escuro'}</button><small>Offline · Porta 14000</small></div></aside><div className="content"><header className="topbar"><div className="global-search"><span>⌕</span><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Pesquisar cliente, projeto, visita, IP, MAC..."/>{results.length>0&&<div className="search-results">{results.map(result=><button key={`${result.type}-${result.id}`} onClick={()=>{onPage(result.type==='device'?'inventory':result.type==='technical_visit'?'visits':'clients');setQuery('');setResults([])}}><b>{result.title}</b><span>{result.type} · {result.subtitle||'sem detalhe'}</span></button>)}</div>}</div><span className="local-badge">● LOCAL</span></header><main className="page">{children}</main></div></div>
}

export function PageHeader({eyebrow,title,description,actions}:{eyebrow?:string;title:string;description?:string;actions?:ReactNode}) { return <div className="page-header"><div><p>{eyebrow}</p><h1>{title}</h1>{description&&<span>{description}</span>}</div><div className="header-actions">{actions}</div></div> }
export function Panel({title,actions,children,className='' }:{title?:string;actions?:ReactNode;children:ReactNode;className?:string}) { return <section className={`panel ${className}`}>{(title||actions)&&<header><h2>{title}</h2><div>{actions}</div></header>}<div className="panel-body">{children}</div></section> }
export function Modal({title,onClose,children,wide=false}:{title:string;onClose:()=>void;children:ReactNode;wide?:boolean}) { useEffect(()=>{const close=(event:KeyboardEvent)=>event.key==='Escape'&&onClose();window.addEventListener('keydown',close);return()=>window.removeEventListener('keydown',close)},[onClose]);return <div className="modal-backdrop" onMouseDown={onClose}><div className={`modal ${wide?'modal-wide':''}`} role="dialog" aria-modal="true" onMouseDown={e=>e.stopPropagation()}><header><h2>{title}</h2><button className="icon-button" onClick={onClose}>×</button></header>{children}</div></div> }
export function Notice({message,onClose}:{message:string;onClose:()=>void}) { return message?<div className="notice"><span>{message}</span><button onClick={onClose}>×</button></div>:null }
export function Status({value}:{value:string}) { const labels:Record<string,string>={online:'✓ Online',offline:'× Offline',unknown:'? Desconhecido',new:'+ Novo',changed:'! Alterado',running:'↻ Executando',completed:'✓ Concluído',cancelled:'× Cancelado',queued:'… Na fila',draft:'Rascunho',scheduled:'Agendada',in_progress:'Em andamento'};return <span className={`status status-${value}`}>{labels[value]??value}</span> }
export function Empty({children}:{children:ReactNode}) { return <div className="empty-state"><span>◇</span><p>{children}</p></div> }
export function Field({label,children}:{label:string;children:ReactNode}) { return <label className="field"><span>{label}</span>{children}</label> }
export function ConfirmButton({message,onConfirm,children,className='danger'}:{message:string;onConfirm:()=>void;children:ReactNode;className?:string}) { return <button type="button" className={className} onClick={()=>window.confirm(message)&&onConfirm()}>{children}</button> }
