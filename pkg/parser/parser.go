// Package parser helps to convert VLESS links into Xray configs.
package parser

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/odlev/xray-cli/pkg/entity"
)

const (
	protocolVless = "vless"
)

// Parse builds an Xray config from a VLESS link and allows overriding the
// socks inbound port.
func Parse(link string, socksPort int) (*entity.Config, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse link: %w", err)
	}

	switch {
	case strings.EqualFold(u.Scheme, protocolVless):
		return parseVless(u, socksPort)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", u.Scheme)
	}
}

func parseVless(u *url.URL, socksPort int) (*entity.Config, error) {
	if u.User == nil || u.User.Username() == "" {
		return nil, fmt.Errorf("empty uuid")
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing host")
	}

	portRaw := u.Port()
	if portRaw == "" {
		return nil, fmt.Errorf("missing port")
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	params := u.Query()
	network := params.Get("type")
	if network == "" {
		network = params.Get("network")
	}
	if network == "" {
		network = "xhttp"
	}

	security := params.Get("security")
	if security == "" {
		security = "reality"
	}

	mode := params.Get("mode")
	if mode == "" {
		mode = "auto"
	}

	cfg := entity.Config{
		API: entity.API{
			Services: []string{"ReflectionService", "HandlerService", "LoggerService", "StatsService"},
			Tag:      "xray-cli_API",
		},
		Inbounds: []entity.Inbound{
			// {
			// 	Listen:   "127.0.0.1",
			// 	Port:     8889,
			// 	Protocol: "http",
			// 	Settings: map[string]any{"allowTransparent": true, "timeout": 300},
			// 	Sniffing: entity.Sniffing{
			// 		Enabled:      true,
			// 		DestOverride: []string{"http", "tls", "quic"},
			// 		RouteOnly:    true,
			// 	},
			// 	Tag: "http_in",
			// },
			{
				Listen:   "127.0.0.1",
				Port:     socksPort,
				Protocol: "socks",
				Settings: map[string]any{
					"auth": "noauth", 
					"ip": "127.0.0.1", 
					"udp": true,
				},
				Sniffing: entity.Sniffing{
					Enabled:      true,
					DestOverride: []string{"http", "tls", "quic"},
					RouteOnly:    true,
				},
				Tag: "socks",
			},
		},
		Log: entity.Log{Loglevel: "warning"},
		Outbounds: []entity.Outbound{
			{
				Protocol: protocolVless,
				Settings: entity.OutboundSettings{
					Vnext: []entity.Vnext{
						{
							Address: host,
							Port:    port,
							Users: []entity.User{
								{Encryption: "none", Flow: "", ID: u.User.Username()},
							},
						},
					},
				},
				StreamSettings: entity.StreamSettings{
					Network:  network,
					Security: security,
					RealitySettings: &entity.RealitySettings{
						Network:       network,
						Show:          false,
						Fingerprint:   params.Get("fp"),
						AllowInsecure: true,
						PublicKey:     params.Get("pbk"),
						ShortID:       params.Get("sid"),
						SpiderX:       params.Get("spx"),
						ServerName:    params.Get("sni"),
						Dest:          params.Get("sni") + ":443",
					},
					XhttpSettings: &entity.XhttpSettings{
						Path: params.Get("path"),
						Host: params.Get("host"),
						Mode: mode,
					},
					XtlsSettings: &entity.XtlsSettings{DisableSystemRoot: false},
				},
				Tag: "proxy",
			},
			{
				Protocol:    "freedom",
				SendThrough: "0.0.0.0",
				Settings: entity.OutboundSettings{
					DomainStrategy: "AsIs",
					Redirect:       ":0",
				},
				StreamSettings: entity.StreamSettings{},
				Tag:            "DIRECT",
			},
			{
				Protocol:    "blackhole",
				SendThrough: "0.0.0.0",
				Settings: entity.OutboundSettings{
					Response: map[string]any{"type": "none"},
				},
				StreamSettings: entity.StreamSettings{},
				Tag:            "BLACKHOLE",
			},
		},
		Policy: entity.Policy{System: map[string]bool{
			"statsInboundDownlink":  true,
			"statsInboundUplink":    true,
			"statsOutboundDownlink": true,
			"statsOutboundUplink":   true,
		}},
		Routing: entity.Routing{
			DomainMatcher:  "mph",
			DomainStrategy: "AsIs",
			Rules: []map[string]any{
				{
					"inboundTag":  []string{"XRayGUI_API_inBOUND"},
					"outboundTag": "XRayGUI_API",
					"type":        "field",
				},
				{
					"ip":          []string{"geoip:private"},
					"outboundTag": "DIRECT",
					"type":        "field",
				},
				{
					"ip":          []string{"geoip:by"},
					"outboundTag": "DIRECT",
					"type":        "field",
				},
				{
					"domain":      []string{"by"},
					"outboundTag": "DIRECT",
					"type":        "field",
				},
			},
		},
		Stats: map[string]any{},
	}

	return &cfg, nil
}

func findFreeUDPPort(startPort int) int {
	udpPort := startPort
	for i := 0; i < 10; i++ {
		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+strconv.Itoa(udpPort))
		if err != nil {
			udpPort++
			continue
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			udpPort++
			continue
		}
		conn.Close()

		return udpPort
	}
	return 0
}
