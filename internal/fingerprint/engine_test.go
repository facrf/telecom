package fingerprint

import "testing"

func TestIdentifyCamera(t *testing.T) {
	f := Identify(Input{MAC: "BC:AD:28:00:00:01", Hostname: "camera-portao", Ports: []int{80, 554}})
	if f.Manufacturer != "Hikvision" || f.DeviceType != "Câmera IP" || f.Confidence < .9 {
		t.Fatalf("unexpected fingerprint: %#v", f)
	}
}

func TestIdentifyONVIFResponse(t *testing.T) {
	result := Identify(Input{Methods: []string{"onvif"}, Banner: "Scopes hardware/HIKVISION name/camera"})
	if result.DeviceType != "Câmera IP" || result.Manufacturer != "Hikvision" || result.Confidence < .7 {
		t.Fatalf("fingerprint inesperado: %#v", result)
	}
}
