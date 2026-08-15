package scanner

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strings"
)

const DefaultHostLimit = 4096

type TargetRange struct {
	Start netip.Addr
	End   netip.Addr
	Count uint64
}

func ParseTarget(value string, limit uint64, allowPublic bool) (TargetRange, error) {
	value = strings.TrimSpace(value)
	if limit == 0 {
		limit = DefaultHostLimit
	}
	if strings.Contains(value, "/") {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return TargetRange{}, fmt.Errorf("CIDR inválido")
		}
		prefix = prefix.Masked()
		if !allowPublic && !prefix.Addr().IsPrivate() {
			return TargetRange{}, fmt.Errorf("somente redes privadas são permitidas")
		}
		bits := prefix.Bits()
		totalBits := prefix.Addr().BitLen()
		hostBits := totalBits - bits
		if hostBits > 20 {
			return TargetRange{}, fmt.Errorf("rede excede o limite de hosts")
		}
		count := uint64(1) << hostBits
		if count > limit {
			return TargetRange{}, fmt.Errorf("rede excede o limite de %d hosts", limit)
		}
		start := prefix.Addr()
		var end netip.Addr
		if start.Is4() {
			b := start.As4()
			ipVal := binary.BigEndian.Uint32(b[:]) + uint32(count-1)
			var endBytes [4]byte
			binary.BigEndian.PutUint32(endBytes[:], ipVal)
			end = netip.AddrFrom4(endBytes)
		} else {
			end = start
			for i := uint64(1); i < count; i++ {
				end = end.Next()
			}
		}
		return TargetRange{Start: start, End: end, Count: count}, nil
	}
	if !strings.Contains(value, "-") {
		addr, err := netip.ParseAddr(value)
		if err != nil {
			return TargetRange{}, fmt.Errorf("endereço IP ou faixa inválida")
		}
		if !allowPublic && !addr.IsPrivate() {
			return TargetRange{}, fmt.Errorf("somente redes privadas são permitidas")
		}
		return TargetRange{Start: addr, End: addr, Count: 1}, nil
	}
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return TargetRange{}, fmt.Errorf("faixa IP inválida")
	}
	start, err := netip.ParseAddr(strings.TrimSpace(parts[0]))
	if err != nil {
		return TargetRange{}, fmt.Errorf("IP inicial inválido")
	}
	end, err := netip.ParseAddr(strings.TrimSpace(parts[1]))
	if err != nil || start.BitLen() != end.BitLen() || end.Compare(start) < 0 {
		return TargetRange{}, fmt.Errorf("IP final inválido")
	}
	if !allowPublic && (!start.IsPrivate() || !end.IsPrivate()) {
		return TargetRange{}, fmt.Errorf("somente redes privadas são permitidas")
	}
	count := uint64(0)
	for current := start; ; current = current.Next() {
		count++
		if count > limit {
			return TargetRange{}, fmt.Errorf("faixa excede o limite de %d hosts", limit)
		}
		if current == end {
			break
		}
	}
	return TargetRange{Start: start, End: end, Count: count}, nil
}
