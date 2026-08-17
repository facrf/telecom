import { ReactNode, useEffect, useRef, useState } from 'react'
import { api, items } from './api'
import { Project } from './types'

export type Page = 'dashboard'|'clients'|'visits'|'inventory'|'scanner'|'topology'|'documentation'|'transfer'|'settings'
export type SearchResult={type:string;id:string;title:string;subtitle:string;projectId?:string}
const navigation: {page:Page;label:string;icon:string;group:'work'|'tools'|'admin'}[] = [
  {page:'dashboard',label:'Início',icon:'▦',group:'work'},{page:'visits',label:'Atendimentos',icon:'✓',group:'work'},
  {page:'clients',label:'Clientes e projetos',icon:'◉',group:'work'},{page:'inventory',label:'Equipamentos',icon:'▣',group:'work'},
  {page:'scanner',label:'Scanner',icon:'⌁',group:'tools'},{page:'topology',label:'Topologia',icon:'⌘',group:'tools'},
  {page:'documentation',label:'Documentação',icon:'▤',group:'tools'},{page:'transfer',label:'Dados e backups',icon:'⇄',group:'admin'},
  {page:'settings',label:'Configurações',icon:'⚙',group:'admin'},
]

export function Shell({page,onPage,theme,onTheme,activeProjectId,onProject,onSearchResult,children}:{page:Page;onPage:(page:Page)=>void;theme:string;onTheme:()=>void;activeProjectId:string;onProject:(id:string)=>void;onSearchResult:(result:SearchResult)=>void;children:ReactNode}) {
  const [query,setQuery]=useState('')
  const [results,setResults]=useState<SearchResult[]>([])
  const [projects,setProjects]=useState<Project[]>([])
  const [menuOpen,setMenuOpen]=useState(false)
  useEffect(()=>{void api<{items:Project[]}>('/projects').then(value=>{const list=items(value);setProjects(list);if(activeProjectId&&!list.some(project=>project.id===activeProjectId))onProject('')}).catch(()=>setProjects([]))},[])
  useEffect(()=>{if(query.trim().length<2){setResults([]);return};const timer=setTimeout(()=>void api<{items:SearchResult[]}>(`/search?q=${encodeURIComponent(query)}`).then(value=>setResults(items(value))).catch(()=>setResults([])),250);return()=>clearTimeout(timer)},[query])
  const go=(target:Page)=>{onPage(target);setMenuOpen(false)}
  const resultLabel:Record<string,string>={client:'Cliente',project:'Projeto',device:'Equipamento',technical_visit:'Visita'}
  return <div className={`shell ${menuOpen?'menu-open':''}`}>
    <aside className="sidebar"><div className="brand"><span className="brand-mark">T</span><div>TELECOM<small>GESTÃO LOCAL</small></div></div><nav>{navigation.map((item,index)=><div key={item.page} className="nav-item">{index>0&&navigation[index-1].group!==item.group&&<small className="nav-group">{item.group==='tools'?'Ferramentas':'Administração'}</small>}<button className={page===item.page?'nav-active':''} onClick={()=>go(item.page)}><span>{item.icon}</span>{item.label}</button></div>)}</nav><div className="sidebar-footer"><button onClick={onTheme}>{theme==='dark'?'☀ Modo claro':'◐ Modo escuro'}</button><small>Offline · Porta 14000</small></div></aside>
    {menuOpen&&<button className="menu-scrim" aria-label="Fechar menu" onClick={()=>setMenuOpen(false)}/>}<div className="content"><header className="topbar"><button className="mobile-menu" aria-label="Abrir menu" onClick={()=>setMenuOpen(true)}>☰</button><div className="project-context"><small>Projeto ativo</small><select aria-label="Projeto ativo" value={activeProjectId} onChange={e=>onProject(e.target.value)}><option value="">Todos os projetos</option>{projects.map(project=><option key={project.id} value={project.id}>{project.name}</option>)}</select></div><div className="global-search"><span>⌕</span><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Pesquisar cliente, visita, IP, MAC…"/>{results.length>0&&<div className="search-results">{results.map(result=><button key={`${result.type}-${result.id}`} onClick={()=>{onSearchResult(result);setQuery('');setResults([])}}><b>{result.title}</b><span>{resultLabel[result.type]||result.type} · {result.subtitle||'sem detalhe'}</span></button>)}</div>}</div><span className="local-badge">● LOCAL</span></header><main className="page">{children}</main></div>
    <nav className="mobile-bottom-nav">{navigation.filter(item=>['dashboard','visits','inventory','scanner'].includes(item.page)).map(item=><button key={item.page} className={page===item.page?'nav-active':''} onClick={()=>go(item.page)}><span>{item.icon}</span><small>{item.label}</small></button>)}</nav>
  </div>
}

export function PageHeader({eyebrow,title,description,actions}:{eyebrow?:string;title:string;description?:string;actions?:ReactNode}) { return <div className="page-header"><div><p>{eyebrow}</p><h1>{title}</h1>{description&&<span>{description}</span>}</div><div className="header-actions">{actions}</div></div> }
export function Panel({title,actions,children,className='' }:{title?:string;actions?:ReactNode;children:ReactNode;className?:string}) { return <section className={`panel ${className}`}>{(title||actions)&&<header><h2>{title}</h2><div>{actions}</div></header>}<div className="panel-body">{children}</div></section> }
export function Modal({title,onClose,children,wide=false,confirmClose=false}:{title:string;onClose:()=>void;children:ReactNode;wide?:boolean;confirmClose?:boolean}) {
  const dialog=useRef<HTMLDivElement>(null)
  const closeRef=useRef(onClose)
  const confirmRef=useRef(confirmClose)
  closeRef.current=onClose
  confirmRef.current=confirmClose
  const requestClose=()=>{if(!confirmRef.current||window.confirm('Existem alterações não salvas. Deseja fechar mesmo assim?'))closeRef.current()}
  useEffect(()=>{const focusable=()=>Array.from(dialog.current?.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),select:not(:disabled),textarea:not(:disabled),a[href]')||[]);focusable()[0]?.focus();const keydown=(event:KeyboardEvent)=>{if(event.key==='Escape'){requestClose();return}if(event.key==='Tab'){const values=focusable();if(!values.length)return;const first=values[0],last=values[values.length-1];if(event.shiftKey&&document.activeElement===first){event.preventDefault();last.focus()}else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus()}}};window.addEventListener('keydown',keydown);return()=>window.removeEventListener('keydown',keydown)},[])
  return <div className="modal-backdrop" onMouseDown={requestClose}><div ref={dialog} className={`modal ${wide?'modal-wide':''}`} role="dialog" aria-modal="true" aria-labelledby="modal-title" onMouseDown={e=>e.stopPropagation()}><header><h2 id="modal-title">{title}</h2><button type="button" aria-label="Fechar" className="icon-button" onClick={requestClose}>×</button></header>{children}</div></div>
}
export function Notice({message,onClose}:{message:string;onClose:()=>void}) { return message?<div className="notice"><span>{message}</span><button aria-label="Fechar aviso" onClick={onClose}>×</button></div>:null }
export function Status({value}:{value:string}) { const labels:Record<string,string>={online:'✓ Online',offline:'× Offline',unknown:'? Desconhecido',new:'+ Novo',changed:'! Alterado',running:'↻ Executando',completed:'✓ Concluído',cancelled:'× Cancelado',queued:'… Na fila',draft:'Rascunho',scheduled:'Agendada',in_progress:'Em andamento'};return <span className={`status status-${value}`}>{labels[value]??value}</span> }
export function Empty({children}:{children:ReactNode}) { return <div className="empty-state"><span>◇</span><p>{children}</p></div> }
export function Field({label,children,error}:{label:string;children:ReactNode;error?:string}) { return <label className={`field ${error?'field-error':''}`}><span>{label}</span>{children}{error&&<small role="alert">{error}</small>}</label> }
export function ConfirmButton({message,onConfirm,children,className='danger'}:{message:string;onConfirm:()=>void;children:ReactNode;className?:string}) { return <button type="button" className={className} onClick={()=>window.confirm(message)&&onConfirm()}>{children}</button> }
