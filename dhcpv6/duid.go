package dhcpv6

import (
	"encoding/binary"
	"fmt"
	"github.com/insomniacslk/dhcp/iana"
)

type DuidType uint16

const (
	_ DuidType = iota
	DUID_LLT
	DUID_EN
	DUID_LL
)

var DuidTypeToString = map[DuidType]string{
	DUID_LL:  "DUID-LL",
	DUID_LLT: "DUID-LLT",
	DUID_EN:  "DUID-EN",
}

type Duid struct {
	Type                 DuidType
	HwType               iana.HwTypeType // for DUID-LLT and DUID-LL. Ignored otherwise. RFC 826
	Time                 uint32          // for DUID-LLT. Ignored otherwise
	LinkLayerAddr        []byte
	EnterpriseNumber     uint32 // for DUID-EN. Ignored otherwise
	EnterpriseIdentifier []byte // for DUID-EN. Ignored otherwise
}

func (d *Duid) Length() int {
	if d.Type == DUID_LLT {
		return 8 + len(d.LinkLayerAddr)
	} else if d.Type == DUID_LL {
		return 4 + len(d.LinkLayerAddr)
	} else if d.Type == DUID_EN {
		return 6 + len(d.EnterpriseIdentifier)
	}
	panic(fmt.Sprintf("Unknown DUID type: %v", d.Type))
}

func (d *Duid) ToBytes() []byte {
	if d.Type == DUID_LLT {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint16(buf[0:2], uint16(d.Type))
		binary.BigEndian.PutUint16(buf[2:4], uint16(d.HwType))
		binary.BigEndian.PutUint32(buf[4:8], d.Time)
		return append(buf, d.LinkLayerAddr...)
	} else if d.Type == DUID_LL {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint16(buf[0:2], uint16(d.Type))
		binary.BigEndian.PutUint16(buf[2:4], uint16(d.HwType))
		return append(buf, d.LinkLayerAddr...)
	} else if d.Type == DUID_EN {
		buf := make([]byte, 6)
		binary.BigEndian.PutUint16(buf[0:2], uint16(d.Type))
		binary.BigEndian.PutUint32(buf[2:6], d.EnterpriseNumber)
		return append(buf, d.EnterpriseIdentifier...)
	}
	panic(fmt.Sprintf("Unknown DUID type: %v", d.Type))
}

func (d *Duid) String() string {
	dtype := DuidTypeToString[d.Type]
	if dtype == "" {
		dtype = "Unknown"
	}
	hwtype := iana.HwTypeToString[d.HwType]
	if hwtype == "" {
		hwtype = "Unknown"
	}
	var hwaddr string
	if d.HwType == iana.HwTypeEthernet {
		for _, b := range d.LinkLayerAddr {
			hwaddr += fmt.Sprintf("%02x:", b)
		}
		if len(hwaddr) > 0 && hwaddr[len(hwaddr)-1] == ':' {
			hwaddr = hwaddr[:len(hwaddr)-1]
		}
	}
	return fmt.Sprintf("DUID{type=%v hwtype=%v hwaddr=%v}", dtype, hwtype, hwaddr)
}

func DuidFromBytes(data []byte) (*Duid, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("Invalid DUID: shorter than 2 bytes")
	}
	d := Duid{}
	d.Type = DuidType(binary.BigEndian.Uint16(data[0:2]))
	if d.Type == DUID_LLT {
		if len(data) < 8 {
			return nil, fmt.Errorf("Invalid DUID-LLT: shorter than 8 bytes")
		}
		d.HwType = iana.HwTypeType(binary.BigEndian.Uint16(data[2:4]))
		d.Time = binary.BigEndian.Uint32(data[4:8])
		d.LinkLayerAddr = data[8:]
	} else if d.Type == DUID_LL {
		if len(data) < 4 {
			return nil, fmt.Errorf("Invalid DUID-LL: shorter than 4 bytes")
		}
		d.HwType = iana.HwTypeType(binary.BigEndian.Uint16(data[2:4]))
		d.LinkLayerAddr = data[4:]
	} else if d.Type == DUID_EN {
		if len(data) < 6 {
			return nil, fmt.Errorf("Invalid DUID-EN: shorter than 6 bytes")
		}
		d.EnterpriseNumber = binary.BigEndian.Uint32(data[2:6])
		d.EnterpriseIdentifier = data[6:]
	}
	return &d, nil
}
