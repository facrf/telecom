export type Client = { id: string; name: string; legalName: string; document: string; phone: string; email: string; contactName: string; address: string; city: string; state: string; postalCode: string; description: string; notes: string }
export type Project = { id: string; clientId: string; name: string; description: string; location: string; address: string; localContact: string; notes: string }
export type Category = { id: string; name: string }
export type Device = { id: string; projectId: string; name: string; description: string; categoryId: string; manufacturer: string; model: string; serialNumber: string; hostname: string; vlan: string; location: string; room: string; rack: string; rackPosition: string; operatingSystem: string; firmware: string; adminUrl: string; status: string; notes: string }
export type Address = { id: string; deviceId: string; type: 'ipv4'|'ipv6'|'mac'; address: string; interface: string; vlan: string; primary: boolean }
export type Tag = { id: string; name: string; color: string }
export type Attachment = { id: string; filename: string; mimeType: string; size: number; description: string; createdAt: string }
export type Scan = { id: string; projectId: string; network: string; status: string; hostsScanned: number; hostsFound: number; startedAt: string; finishedAt: string }
export type ScanHost = { ip: string; mac: string; hostname: string; status: string; discoveryMethod: string; discoveryMethods: string[]; manufacturer: string; deviceType: string; confidence: number; evidence: {source:string;detail:string;weight:number}[] }
export type ScanChange = { Type?: string; type?: string; Subject?: string; subject?: string; Previous?: string; previous?: string; Current?: string; current?: string }
export type PortScan = { id: string; deviceId: string; mode: string; ports: string; status: string; startedAt: string; finishedAt: string }
export type PortResult = { port: number; protocol: string; state: string; service: string; product: string; version: string; banner: string; confidence: number }
export type Diagram = { id: string; projectId: string; name: string; description: string }
export type DiagramNode = { id: string; diagramId: string; deviceId: string; label: string; x: number; y: number; width: number; height: number; styleJson: string }
export type DiagramEdge = { id: string; diagramId: string; sourceNodeId: string; targetNodeId: string; name: string; description: string; type: string; sourceInterface: string; targetInterface: string; speed: string; vlan: string; technology: string; color: string; lineStyle: string; notes: string }
export type NetworkDocument = { id: string; projectId: string; title: string; responsible: string; generalDescription: string; internetWan: string; lan: string; vlans: string; wifi: string; cctv: string; telephony: string; servers: string; racks: string; cabling: string; fiber: string; links: string; power: string; procedures: string; notes: string; freeText: string }
export type Summary = { clients:number;projects:number;devices:number;online:number;offline:number;cameras:number;routers:number;switches:number;servers:number }
export type AuditEntry = { id:string;action:string;entityType:string;entityId:string;createdAt:string }
export type TechnicalVisit = {
  id:string;projectId:string;clientId:string;clientName:string;projectName:string;protocol:string;title:string;
  visitType:string;status:string;result:string;scheduledAt:string;startedAt:string;finishedAt:string;durationMinutes:number;
  responsibleTechnician:string;requester:string;localContact:string;requestDescription:string;initialSituation:string;
  diagnosis:string;workSummary:string;recommendations:string;pendingSummary:string;customerNotes:string;internalNotes:string;
  requiresReturn:boolean;returnReason:string;suggestedReturnAt:string;createdAt:string;updatedAt:string
}
export type VisitDeviceItem = {id:string;technicalVisitId:string;deviceId:string;deviceName:string;deviceCategory:string;role:string;notes:string;createdAt:string;updatedAt:string}
export type VisitServiceItem = {id:string;technicalVisitId:string;description:string;category:string;deviceId:string;deviceName:string;performedAt:string;technician:string;notes:string;order:number;createdAt:string;updatedAt:string}
export type VisitChecklistItem = {id:string;technicalVisitId:string;text:string;status:string;notes:string;order:number;createdAt:string;updatedAt:string}
export type VisitMaterialItem = {id:string;technicalVisitId:string;quantity:number;unit:string;description:string;brand:string;model:string;notes:string;createdAt:string;updatedAt:string}
export type VisitPendingItem = {id:string;technicalVisitId:string;description:string;priority:string;responsible:string;dueAt:string;status:string;createdAt:string;updatedAt:string}
