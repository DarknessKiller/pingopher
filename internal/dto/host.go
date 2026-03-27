package dto

import (
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/segmentio/ksuid"
)

type CreateHostRequest struct {
	Name                string       `json:"name" binding:"required"`
	Protocol            string       `json:"protocol" binding:"required,oneof=http https"`
	HostURL             string       `json:"hostUrl" binding:"required"`
	Port                uint16       `json:"port" binding:"omitempty,numeric,gte=1,lte=65535"`
	TLS                 tlsRequest   `json:"tls" binding:"omitempty"`
	DNS                 []dnsRequest `json:"dns" binding:"omitempty,dive"`
	PingInterval        uint16       `json:"pingInterval" binding:"required,numeric"`
	FailThreshold       uint16       `json:"failThreshold" binding:"required,numeric"`
	AcceptedStatusCodes []string     `json:"acceptedStatusCodes" binding:"required"`
}

type dnsRequest struct {
	Name     string `json:"name" binding:"required"`
	IP       string `json:"ip" binding:"required,ipv4|ipv6"`
	Port     uint16 `json:"port" binding:"required,numeric,gte=1,lte=65535"`
	Protocol string `json:"protocol" binding:"required,oneof=udp tcp"`
}

type tlsRequest struct {
	NoVerify *bool `json:"no_verify" binding:"omitempty"`
}

func (r CreateHostRequest) ToModel() *model.Host {
	var dns []model.DNS
	if r.DNS != nil {
		dns = make([]model.DNS, len(r.DNS))
		for i, d := range r.DNS {
			dns[i] = model.DNS{
				Name:     d.Name,
				IP:       d.IP,
				Port:     d.Port,
				Protocol: d.Protocol,
			}
		}
	}

	var tls model.TLS
	if r.TLS.NoVerify != nil {
		tls = model.TLS{
			NoVerify: r.TLS.NoVerify,
		}
	}

	return &model.Host{
		Name:                r.Name,
		Protocol:            r.Protocol,
		HostURL:             r.HostURL,
		Port:                &r.Port,
		TLS:                 tls,
		DNS:                 dns,
		PingInterval:        r.PingInterval,
		FailThreshold:       r.FailThreshold,
		AcceptedStatusCodes: r.AcceptedStatusCodes,
		Status:              model.HostStatusUnknown,
	}
}

type Host struct {
	ID                  ksuid.KSUID `json:"id"`
	Name                string      `json:"name"`
	Protocol            string      `json:"protocol"`
	HostURL             string      `json:"hostUrl"`
	Port                uint16      `json:"port"`
	TLS                 model.TLS   `json:"tls"`
	DNS                 []model.DNS `json:"dns"`
	PingInterval        uint16      `json:"pingInterval"`
	FailThreshold       uint16      `json:"failThreshold"`
	AcceptedStatusCodes []string    `json:"acceptedStatusCodes"`
	Status              string      `json:"status"`
}

func ToHost(host *model.Host) Host {
	return Host{
		ID:                  host.ID,
		Name:                host.Name,
		Protocol:            host.Protocol,
		HostURL:             host.HostURL,
		Port:                *host.Port,
		TLS:                 host.TLS,
		DNS:                 host.DNS,
		PingInterval:        host.PingInterval,
		FailThreshold:       host.FailThreshold,
		AcceptedStatusCodes: host.AcceptedStatusCodes,
		Status:              string(host.Status),
	}
}

type Hosts struct {
	Hosts []Host `json:"hosts"`
}

func ToAllHosts(hosts []model.Host) Hosts {
	res := make([]Host, len(hosts))

	for i := range hosts {
		res[i] = ToHost(&hosts[i])
	}

	return Hosts{Hosts: res}
}

type UpdateHostRequest struct {
	Name                string       `json:"name" binding:"omitempty"`
	Protocol            string       `json:"protocol" binding:"omitempty,oneof=http https"`
	HostURL             string       `json:"hostUrl" binding:"omitempty"`
	Port                uint16       `json:"port" binding:"omitempty,numeric,gte=1,lte=65535"`
	TLS                 tlsRequest   `json:"tls" binding:"omitempty"`
	DNS                 []dnsRequest `json:"dns" binding:"omitempty,dive"`
	PingInterval        uint16       `json:"pingInterval" binding:"omitempty,numeric"`
	FailThreshold       uint16       `json:"failThreshold" binding:"omitempty,numeric"`
	AcceptedStatusCodes []string     `json:"acceptedStatusCodes" binding:"required"`
}

func (r UpdateHostRequest) ToModel() *model.Host {
	var dns []model.DNS
	if r.DNS != nil {
		dns = make([]model.DNS, len(r.DNS))
		for i, d := range r.DNS {
			dns[i] = model.DNS{
				Name:     d.Name,
				IP:       d.IP,
				Port:     d.Port,
				Protocol: d.Protocol,
			}
		}
	}

	var tls model.TLS
	if r.TLS.NoVerify != nil {
		tls = model.TLS{
			NoVerify: r.TLS.NoVerify,
		}
	}

	return &model.Host{
		Name:                r.Name,
		Protocol:            r.Protocol,
		HostURL:             r.HostURL,
		Port:                &r.Port,
		TLS:                 tls,
		DNS:                 dns,
		PingInterval:        r.PingInterval,
		FailThreshold:       r.FailThreshold,
		AcceptedStatusCodes: r.AcceptedStatusCodes,
	}
}
