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
	Confidence   float64    `json:"confidence"`
	Evidence     []Evidence `json:"evidence"`
}
type Input struct {
	MAC, Hostname string
	Ports         []int
	Methods       []string
	Banner        string
}

func Identify(input Input) DeviceFingerprint {
	f := DeviceFingerprint{DeviceType: "Desconhecido"}
	mac := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(input.MAC, ":", ""), "-", ""))
	ouis := map[string]string{"BCAD28": "Hikvision", "3C2DB7": "Hikvision", "A0F3C1": "Dahua", "9C8E99": "Intelbras", "B0A737": "TP-Link", "001A2B": "AOC", "001B21": "Intelbras"}
	if manufacturer := ouis[mac[:min(6, len(mac))]]; manufacturer != "" {
		f.Manufacturer = manufacturer
		f.Evidence = append(f.Evidence, Evidence{"MAC OUI", "Prefixo MAC associado a " + manufacturer, .45})
	}
	host := strings.ToLower(input.Hostname)
	banner := strings.ToLower(input.Banner)
	manufacturers := map[string]string{"hikvision": "Hikvision", "dahua": "Dahua", "intelbras": "Intelbras", "axis": "Axis", "uniview": "Uniview", "hanwha": "Hanwha", "vigi": "TP-Link VIGI", "bosch": "Bosch"}
	for marker, manufacturer := range manufacturers {
		if strings.Contains(banner, marker) {
			f.Manufacturer = manufacturer
			f.Evidence = append(f.Evidence, Evidence{"Resposta de descoberta", "Identificação compatível com " + manufacturer, .4})
			break
		}
	}
	for _, method := range input.Methods {
		if method == "onvif" {
			f.DeviceType = "Câmera IP"
			f.Evidence = append(f.Evidence, Evidence{"ONVIF", "Resposta WS-Discovery recebida", .4})
		}
	}
	if strings.Contains(host, "camera") || strings.Contains(host, "cam-") {
		f.DeviceType = "Câmera IP"
		f.Evidence = append(f.Evidence, Evidence{"Hostname", "Nome compatível com câmera", .25})
	}
	for _, port := range input.Ports {
		switch port {
		case 554, 8554:
			f.DeviceType = "Câmera IP"
			f.Evidence = append(f.Evidence, Evidence{"Porta", "RTSP detectado", .3})
		case 37777, 34567:
			if f.Manufacturer == "" {
				f.Manufacturer = "Dahua"
			}
			f.DeviceType = "Câmera IP"
			f.Evidence = append(f.Evidence, Evidence{"Porta", "Serviço de CFTV detectado", .4})
		case 161:
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Equipamento de rede"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta", "SNMP detectado", .2})
		case 22:
			if f.DeviceType == "Desconhecido" {
				f.DeviceType = "Servidor ou equipamento de rede"
			}
			f.Evidence = append(f.Evidence, Evidence{"Porta", "SSH detectado", .15})
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
