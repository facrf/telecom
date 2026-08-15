package fingerprint

import "strings"

type Evidence struct {
	Source string  `json:"source"`
	Detail string  `json:"detail"`
	Weight float64 `json:"weight"`
}

type DeviceFingerprint struct {
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	DeviceType   string     `json:"deviceType"`
	CategoryID   string     `json:"categoryId"`
	Confidence   float64    `json:"confidence"`
	Evidence     []Evidence `json:"evidence"`
}

type Input struct {
	MAC, Hostname string
	Ports         []int
	Methods       []string
	Banner        string
}

type ouiEntry struct {
	Vendor   string
	DevType  string
	Category string
}

var knownOUIs = map[string]ouiEntry{
	// MikroTik
	"488F5A": {"MikroTik", "Router / Switch", "router"},
	"6C3B6B": {"MikroTik", "Router / Switch", "router"},
	"000C42": {"MikroTik", "Router / Switch", "router"},
	"CC2DE0": {"MikroTik", "Router / Switch", "router"},
	"D401C3": {"MikroTik", "Router / Switch", "router"},
	"E48D8C": {"MikroTik", "Router / Switch", "router"},
	"789A18": {"MikroTik", "Router / Switch", "router"},
	"2C6B7D": {"MikroTik", "Router / Switch", "router"},
	"B869F4": {"MikroTik", "Router / Switch", "router"},

	// Ubiquiti / UniFi
	"002722": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"0418D6": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"24A43C": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"68D79A": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"70A741": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"7483C2": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"ACEC80": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"B4FBE4": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"DC9FDB": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"F09FC2": {"Ubiquiti", "Access Point / Switch", "access-point"},
	"E063DA": {"Ubiquiti", "Access Point / Switch", "access-point"},

	// Cisco
	"00000C": {"Cisco", "Switch / Router", "switch"},
	"000142": {"Cisco", "Switch / Router", "switch"},
	"001E13": {"Cisco", "Switch / Router", "switch"},
	"001DA1": {"Cisco", "Switch / Router", "switch"},
	"001FCA": {"Cisco", "Switch / Router", "switch"},
	"002255": {"Cisco", "Switch / Router", "switch"},
	"002414": {"Cisco", "Switch / Router", "switch"},
	"002698": {"Cisco", "Switch / Router", "switch"},
	"F4CFE2": {"Cisco", "Switch / Router", "switch"},
	"708105": {"Cisco", "Switch / Router", "switch"},
	"58971E": {"Cisco", "Switch / Router", "switch"},

	// TP-Link
	"B0A737": {"TP-Link", "Roteador / Switch", "router"},
	"50C7BF": {"TP-Link", "Roteador / Switch", "router"},
	"E848B8": {"TP-Link", "Roteador / Switch", "router"},
	"14CC20": {"TP-Link", "Roteador / Switch", "router"},
	"54E6FC": {"TP-Link", "Roteador / Switch", "router"},
	"003192": {"TP-Link", "Roteador / Switch", "router"},
	"704F57": {"TP-Link", "Roteador / Switch", "router"},

	// Intelbras
	"001B21": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},
	"9C8E99": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},
	"B0C554": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},
	"3C8375": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},
	"145F94": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},
	"806C1B": {"Intelbras", "Dispositivo Telecom / CFTV", "other"},

	// Hikvision
	"BCAD28": {"Hikvision", "Câmera IP", "ip-camera"},
	"3C2DB7": {"Hikvision", "Câmera IP", "ip-camera"},
	"C056E3": {"Hikvision", "Câmera IP", "ip-camera"},
	"4419B6": {"Hikvision", "Câmera IP", "ip-camera"},
	"2857BE": {"Hikvision", "Câmera IP", "ip-camera"},
	"0018AE": {"Hikvision", "Câmera IP", "ip-camera"},
	"40167E": {"Hikvision", "Câmera IP", "ip-camera"},

	// Dahua
	"A0F3C1": {"Dahua", "Câmera IP", "ip-camera"},
	"38AF29": {"Dahua", "Câmera IP", "ip-camera"},
	"E0508B": {"Dahua", "Câmera IP", "ip-camera"},
	"4C11BF": {"Dahua", "Câmera IP", "ip-camera"},
	"9002A9": {"Dahua", "Câmera IP", "ip-camera"},

	// Axis
	"ACCC8E": {"Axis Communications", "Câmera IP", "ip-camera"},
	"00408C": {"Axis Communications", "Câmera IP", "ip-camera"},

	// Uniview
	"D89B3B": {"Uniview", "Câmera IP", "ip-camera"},
	"5C260A": {"Uniview", "Câmera IP", "ip-camera"},

	// GPON / OLT / ONU (Huawei, ZTE, Fiberhome, Datacom)
	"00E0FC": {"Huawei", "ONU / OLT", "onu"},
	"04258B": {"Huawei", "ONU / OLT", "onu"},
	"48437C": {"Huawei", "ONU / OLT", "onu"},
	"D46AA8": {"Huawei", "ONU / OLT", "onu"},
	"F8E71E": {"Huawei", "ONU / OLT", "onu"},
	"002293": {"ZTE", "ONU / OLT", "onu"},
	"00D0D0": {"ZTE", "ONU / OLT", "onu"},
	"146080": {"ZTE", "ONU / OLT", "onu"},
	"94A7B7": {"ZTE", "ONU / OLT", "onu"},
	"000AC7": {"Fiberhome", "ONU / OLT", "onu"},
	"581F28": {"Fiberhome", "ONU / OLT", "onu"},
	"D8B12A": {"Fiberhome", "ONU / OLT", "onu"},
	"0004DF": {"Datacom", "Switch / OLT", "switch"},
	"001C57": {"Datacom", "Switch / OLT", "switch"},

	// Servidores & TI (Dell, HP, VMware)
	"001422": {"Dell", "Servidor", "server"},
	"14FEB5": {"Dell", "Servidor", "server"},
	"B82A72": {"Dell", "Servidor", "server"},
	"001F29": {"HP / HPE", "Servidor / Switch", "server"},
	"3CD92B": {"HP / HPE", "Servidor / Switch", "server"},
	"000C29": {"VMware", "Máquina Virtual / Servidor", "server"},
	"005056": {"VMware", "Máquina Virtual / Servidor", "server"},
	"00155D": {"Hyper-V", "Máquina Virtual / Servidor", "server"},

	// Firewalls & Segurança
	"00090F": {"Fortinet", "Firewall / FortiGate", "firewall"},
	"704C95": {"Fortinet", "Firewall / FortiGate", "firewall"},
	"906C2C": {"Fortinet", "Firewall / FortiGate", "firewall"},
	"00065B": {"SonicWall", "Firewall", "firewall"},

	// Telefonia IP (Grandstream, Yealink)
	"000B82": {"Grandstream", "Telefone IP / ATA", "ip-phone"},
	"001565": {"Yealink", "Telefone IP", "ip-phone"},
	"805EC0": {"Yealink", "Telefone IP", "ip-phone"},

	// Nobreak / UPS (APC / Schneider)
	"00025D": {"APC / Schneider", "UPS / Nobreak", "ups"},
	"000B3B": {"APC / Schneider", "UPS / Nobreak", "ups"},
	"00C0B7": {"APC / Schneider", "UPS / Nobreak", "ups"},

	// Storage / NAS
	"001132": {"Synology", "NAS / Storage", "nas"},
	"00089B": {"QNAP", "NAS / Storage", "nas"},
	"245EBE": {"QNAP", "NAS / Storage", "nas"},

	// IoT & Automação
	"240AC4": {"Espressif", "Dispositivo IoT", "iot"},
	"30AEA4": {"Espressif", "Dispositivo IoT", "iot"},
	"84F3EB": {"Espressif", "Dispositivo IoT", "iot"},
	"A4CF12": {"Espressif", "Dispositivo IoT", "iot"},
}

type vendorMarker struct {
	Keyword      string
	Manufacturer string
	Category     string
	DevType      string
}

var vendorMarkers = []vendorMarker{
	{"mikrotik", "MikroTik", "router", "Router MikroTik"},
	{"routeros", "MikroTik", "router", "Router MikroTik"},
	{"winbox", "MikroTik", "router", "Router MikroTik"},
	{"unifi", "Ubiquiti", "access-point", "Access Point UniFi"},
	{"ubiquiti", "Ubiquiti", "access-point", "Equipamento Ubiquiti"},
	{"airos", "Ubiquiti", "radio", "Rádio / Antena airMAX"},
	{"nanostation", "Ubiquiti", "radio", "Antena NanoStation"},
	{"litebeam", "Ubiquiti", "radio", "Antena LiteBeam"},
	{"powerbeam", "Ubiquiti", "radio", "Antena PowerBeam"},
	{"fortigate", "Fortinet", "firewall", "Firewall FortiGate"},
	{"fortinet", "Fortinet", "firewall", "Firewall Fortinet"},
	{"pfsense", "pfSense", "firewall", "Firewall pfSense"},
	{"opnsense", "OPNsense", "firewall", "Firewall OPNsense"},
	{"sonicwall", "SonicWall", "firewall", "Firewall SonicWall"},
	{"sophos", "Sophos", "firewall", "Firewall Sophos"},
	{"hikvision", "Hikvision", "ip-camera", "Câmera IP Hikvision"},
	{"dahua", "Dahua", "ip-camera", "Câmera IP Dahua"},
	{"intelbras", "Intelbras", "other", "Equipamento Intelbras"},
	{"axis", "Axis Communications", "ip-camera", "Câmera IP Axis"},
	{"uniview", "Uniview", "ip-camera", "Câmera IP Uniview"},
	{"hanwha", "Hanwha Vision", "ip-camera", "Câmera IP Hanwha"},
	{"vigi", "TP-Link VIGI", "ip-camera", "Câmera IP VIGI"},
	{"bosch", "Bosch", "ip-camera", "Câmera IP Bosch"},
	{"grandstream", "Grandstream", "ip-phone", "Telefone IP Grandstream"},
	{"yealink", "Yealink", "ip-phone", "Telefone IP Yealink"},
	{"asterisk", "Asterisk", "ip-phone", "PABX IP Asterisk"},
	{"freepbx", "FreePBX", "ip-phone", "PABX IP FreePBX"},
	{"synology", "Synology", "nas", "NAS Synology DiskStation"},
	{"qnap", "QNAP", "nas", "NAS QNAP Turbo"},
	{"truenas", "TrueNAS", "nas", "Storage TrueNAS"},
	{"proxmox", "Proxmox", "server", "Servidor Proxmox VE"},
	{"esxi", "VMware", "server", "Servidor VMware ESXi"},
	{"gpon", "GPON", "onu", "ONU / ONT Fibra"},
	{"epon", "EPON", "onu", "ONU / ONT Fibra"},
	{"fiberhome", "Fiberhome", "onu", "ONU / OLT Fiberhome"},
	{"datacom", "Datacom", "switch", "Switch / OLT Datacom"},
	{"smart-ups", "APC", "ups", "Nobreak Smart-UPS"},
	{"apc", "APC", "ups", "Nobreak APC"},
	{"cups", "Impressora", "printer", "Servidor de Impressão"},
	{"jetdirect", "HP", "printer", "Impressora de Rede"},
}

func Identify(input Input) DeviceFingerprint {
	f := DeviceFingerprint{DeviceType: "Desconhecido", CategoryID: "other"}

	// Normalização consistente do MAC
	macReplacer := strings.NewReplacer(":", "", "-", "", ".", "", " ", "")
	mac := strings.ToUpper(macReplacer.Replace(input.MAC))
	if len(mac) >= 6 {
		prefix := mac[:6]
		if entry, ok := knownOUIs[prefix]; ok {
			f.Manufacturer = entry.Vendor
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = entry.DevType
				f.CategoryID = entry.Category
			}
			f.Evidence = append(f.Evidence, Evidence{"MAC OUI", "Prefixo MAC associado a " + entry.Vendor, .45})
		}
	}

	host := strings.ToLower(input.Hostname)
	banner := strings.ToLower(input.Banner)

	// Análise determinística de banners
	for _, marker := range vendorMarkers {
		if strings.Contains(banner, marker.Keyword) || strings.Contains(host, marker.Keyword) {
			if f.Manufacturer == "" {
				f.Manufacturer = marker.Manufacturer
			}
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = marker.DevType
				f.CategoryID = marker.Category
			}
			f.Evidence = append(f.Evidence, Evidence{"Identificação de Banner/Hostname", "Assinatura compatível com " + marker.DevType, .45})
			break
		}
	}

	// Métodos de Descoberta
	for _, method := range input.Methods {
		switch method {
		case "onvif":
			f.DeviceType = "Câmera IP"
			f.CategoryID = "ip-camera"
			f.Evidence = append(f.Evidence, Evidence{"ONVIF", "Resposta WS-Discovery ONVIF recebida", .45})
		case "ssdp":
			if strings.Contains(banner, "av:mediaserver") {
				f.DeviceType = "Servidor de Mídia"
				f.CategoryID = "server"
				f.Evidence = append(f.Evidence, Evidence{"SSDP", "MediaServer SSDP detectado", .3})
			}
		case "mdns":
			if strings.Contains(banner, "_printer") || strings.Contains(banner, "_ipp") {
				f.DeviceType = "Impressora de Rede"
				f.CategoryID = "printer"
				f.Evidence = append(f.Evidence, Evidence{"mDNS", "Serviço de impressão anunciado", .4})
			} else if strings.Contains(banner, "_ubnt") {
				f.DeviceType = "Equipamento Ubiquiti UniFi"
				f.CategoryID = "access-point"
				f.Evidence = append(f.Evidence, Evidence{"mDNS", "Serviço UniFi anunciado", .45})
			}
		}
	}

	// Heurísticas por Hostname
	if strings.Contains(host, "camera") || strings.Contains(host, "cam-") || strings.Contains(host, "cftv") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Câmera IP"
			f.CategoryID = "ip-camera"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com câmera", .25})
	} else if strings.Contains(host, "switch") || strings.Contains(host, "sw-") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Switch de Rede"
			f.CategoryID = "switch"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com Switch", .25})
	} else if strings.Contains(host, "router") || strings.Contains(host, "rt-") || strings.Contains(host, "gw-") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Roteador / Gateway"
			f.CategoryID = "router"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com Roteador", .25})
	} else if strings.Contains(host, "ap-") || strings.Contains(host, "wifi") || strings.Contains(host, "unifi") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Access Point Wi-Fi"
			f.CategoryID = "access-point"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com Access Point", .25})
	} else if strings.Contains(host, "srv") || strings.Contains(host, "server") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Servidor"
			f.CategoryID = "server"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com Servidor", .25})
	} else if strings.Contains(host, "print") || strings.Contains(host, "imp-") {
		if f.DeviceType == "Desconhecido" {
			f.DeviceType = "Impressora"
			f.CategoryID = "printer"
		}
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com Impressora", .25})
	}

	// Análise por Portas Abertas
	for _, port := range input.Ports {
		switch port {
		case 8291: // MikroTik Winbox
			if f.Manufacturer == "" {
				f.Manufacturer = "MikroTik"
			}
			f.DeviceType = "Roteador MikroTik (Winbox)"
			f.CategoryID = "router"
			f.Evidence = append(f.Evidence, Evidence{"Porta 8291", "Porta de gerenciamento MikroTik Winbox aberta", .5})
		case 5060, 5061: // SIP VoIP
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = "Telefone IP / PABX VoIP"
				f.CategoryID = "ip-phone"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta SIP", "Protocolo SIP de telefonia IP ativo", .4})
		case 9100, 631, 515: // Impressoras
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = "Impressora de Rede"
				f.CategoryID = "printer"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta Impressão", "Serviço RAW/IPP de impressão ativo", .4})
		case 5000, 5001: // Synology DSM
			if f.Manufacturer == "" {
				f.Manufacturer = "Synology"
			}
			f.DeviceType = "Storage / NAS Synology"
			f.CategoryID = "nas"
			f.Evidence = append(f.Evidence, Evidence{"Porta DSM", "Gerenciamento Synology DSM ativo", .45})
		case 8006: // Proxmox VE
			if f.Manufacturer == "" {
				f.Manufacturer = "Proxmox"
			}
			f.DeviceType = "Servidor de Virtualização Proxmox"
			f.CategoryID = "server"
			f.Evidence = append(f.Evidence, Evidence{"Porta 8006", "Painel Proxmox VE detectado", .45})
		case 10443, 4444: // FortiGate / Sophos
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = "Firewall / UTM"
				f.CategoryID = "firewall"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta Firewall", "Painel de gerência de Firewall ativo", .35})
		case 554, 8554: // RTSP
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = "Câmera IP (RTSP)"
				f.CategoryID = "ip-camera"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta RTSP", "Streaming RTSP ativo", .35})
		case 37777, 34567: // Dahua / Intelbras CFTV
			if f.Manufacturer == "" {
				f.Manufacturer = "Dahua / Intelbras"
			}
			f.DeviceType = "DVR / NVR / Câmera CFTV"
			f.CategoryID = "nvr"
			f.Evidence = append(f.Evidence, Evidence{"Porta CFTV", "Serviço DVR/NVR detectado", .45})
		case 8000: // Hikvision
			if f.Manufacturer == "" {
				f.Manufacturer = "Hikvision"
			}
			if f.DeviceType == "Desconhecido" || f.CategoryID == "other" {
				f.DeviceType = "Dispositivo Hikvision (DVR/NVR/Câmera)"
				f.CategoryID = "nvr"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta 8000", "Porta de serviço Hikvision ativa", .3})
		case 161: // SNMP
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Equipamento de Rede Gerenciável (Switch/Router)"
				f.CategoryID = "switch"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta SNMP", "Agente SNMP detectado", .25})
		case 3389: // RDP
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Servidor / Estação Windows"
				f.CategoryID = "server"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta RDP", "Terminal Server / Área de Trabalho Remota", .3})
		case 445, 139: // SMB
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Servidor / Compartilhamento SMB"
				f.CategoryID = "server"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta SMB", "Serviço de arquivos SMB ativo", .2})
		case 22: // SSH
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Servidor ou Equipamento Linux"
				f.CategoryID = "server"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta SSH", "Terminal seguro SSH ativo", .15})
		}
	}

	for _, e := range f.Evidence {
		f.Confidence += e.Weight
	}
	if f.Confidence > 1 {
		f.Confidence = 1
	}
	return f
}

