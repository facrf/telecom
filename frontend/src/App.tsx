import { lazy, ReactNode, Suspense, useEffect, useState } from 'react'
import { Page, SearchResult, Shell } from './components'
import DashboardPage from './pages/DashboardPage'

const ClientsPage=lazy(()=>import('./pages/ClientsPage'))
const VisitsPage=lazy(()=>import('./pages/VisitsPage'))
const InventoryPage=lazy(()=>import('./pages/InventoryPage'))
const ScannerPage=lazy(()=>import('./pages/ScannerPage'))
const TopologyPage=lazy(()=>import('./pages/TopologyPage'))
const DocumentationPage=lazy(()=>import('./pages/DocumentationPage'))
const TransferPage=lazy(()=>import('./pages/TransferPage'))
const SettingsPage=lazy(()=>import('./pages/SettingsPage'))

type Route={page:Page;id:string;entityType:string}
const pages=new Set<Page>(['dashboard','clients','visits','inventory','scanner','topology','documentation','transfer','settings'])
function readRoute():Route{const[path,query='']=location.hash.split('?');const parts=path.replace(/^#\/?/,'').split('/').filter(Boolean);const page=pages.has(parts[0] as Page)?parts[0] as Page:'dashboard';return{page,id:decodeURIComponent(parts[1]||''),entityType:new URLSearchParams(query).get('type')||''}}

export default function App(){
  const[route,setRoute]=useState<Route>(readRoute)
  const[activeProjectId,setActiveProjectId]=useState(localStorage.getItem('telecom-active-project')||'')
  const[theme,setTheme]=useState(localStorage.getItem('telecom-theme')||'dark')
  useEffect(()=>{document.documentElement.dataset.theme=theme;localStorage.setItem('telecom-theme',theme)},[theme])
  useEffect(()=>{const changed=()=>setRoute(readRoute());window.addEventListener('hashchange',changed);if(!location.hash)location.hash='#/dashboard';return()=>window.removeEventListener('hashchange',changed)},[])
  const chooseProject=(projectId:string)=>{setActiveProjectId(projectId);if(projectId)localStorage.setItem('telecom-active-project',projectId);else localStorage.removeItem('telecom-active-project')}
  const navigate=(page:Page,id='',entityType='')=>{const query=entityType?`?type=${encodeURIComponent(entityType)}`:'';location.hash=`#/${page}${id?`/${encodeURIComponent(id)}`:''}${query}`}
  const openProjectVisits=(projectId:string)=>{chooseProject(projectId);navigate('visits')}
  const openSearchResult=(result:SearchResult)=>{if(result.projectId)chooseProject(result.projectId);if(result.type==='technical_visit')navigate('visits',result.id,result.type);else if(result.type==='device')navigate('inventory',result.id,result.type);else if(result.type==='project'){chooseProject(result.id);navigate('clients',result.id,result.type)}else navigate('clients',result.id,result.type)}
  const content:Record<Page,ReactNode>={
    dashboard:<DashboardPage projectId={activeProjectId} onNavigate={navigate}/>,
    clients:<ClientsPage initialId={route.id} initialType={route.entityType} onActiveProject={chooseProject} onOpenVisits={openProjectVisits}/>,
    visits:<VisitsPage initialProjectId={activeProjectId} initialVisitId={route.page==='visits'?route.id:''}/>,
    inventory:<InventoryPage initialProjectId={activeProjectId} initialDeviceId={route.page==='inventory'?route.id:''}/>,
    scanner:<ScannerPage initialProjectId={activeProjectId}/>,
    topology:<TopologyPage initialProjectId={activeProjectId}/>,
    documentation:<DocumentationPage initialProjectId={activeProjectId}/>,
    transfer:<TransferPage initialProjectId={activeProjectId}/>,
    settings:<SettingsPage theme={theme} onTheme={setTheme}/>
  }
  return <Shell page={route.page} onPage={page=>navigate(page)} theme={theme} onTheme={()=>setTheme(theme==='dark'?'light':'dark')} activeProjectId={activeProjectId} onProject={chooseProject} onSearchResult={openSearchResult}><Suspense fallback={<div className="page-loading">Carregando módulo…</div>}>{content[route.page]}</Suspense></Shell>
}
