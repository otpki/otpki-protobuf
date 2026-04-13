package converters

import (
	"crypto/x509/pkix"
	"fmt"

	x509pb "github.com/otpki/otpki-protobuf/gen/otpki/common/x509/v1"
)

// Helper functions for converting between the standard library and protobuf types

// PKIXNameToProtoName converts a pkix.Name to a protobuf Name
func PKIXNameToProtoName(name pkix.Name) *x509pb.Name {
	protoName := &x509pb.Name{
		Country:            name.Country,
		Organization:       name.Organization,
		OrganizationalUnit: name.OrganizationalUnit,
		Locality:           name.Locality,
		Province:           name.Province,
		StreetAddress:      name.StreetAddress,
		PostalCode:         name.PostalCode,
		SerialNumber:       name.SerialNumber,
		CommonName:         name.CommonName,
	}

	for _, n := range name.Names {
		// Convert additional names (OID type, value pairs)
		if len(n.Type) > 0 {
			oidStr := OIDToString(n.Type)
			if oidStr != "" && n.Value != nil {
				// Convert the value to string (handling different types)
				valueStr := ""
				switch v := n.Value.(type) {
				case string:
					valueStr = v
				case []byte:
					valueStr = string(v)
				default:
					valueStr = fmt.Sprintf("%v", v)
				}

				if valueStr != "" {
					attr := &x509pb.NameAttribute{
						Type:  oidStr,
						Value: nil,
					}
					protoName.Names = append(protoName.Names, attr)
				}
			}
		}
	}

	return protoName
}
