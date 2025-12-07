// Package entity
package entity

type Config struct {
	API       API            `json:"api"`
	Inbounds  []Inbound      `json:"inbounds"`
	Log       Log            `json:"log"`
	Outbounds []Outbound     `json:"outbounds"`
	Policy    Policy         `json:"policy"`
	Routing   Routing        `json:"routing"`
	Stats     map[string]any `json:"stats"`
}

type API struct {
	Services []string `json:"services"`
	Tag      string   `json:"tag"`
}

type Inbound struct {
	Listen   string         `json:"listen"`
	Port     int            `json:"port"`
	Protocol string         `json:"protocol"`
	Settings map[string]any `json:"settings"`
	Sniffing Sniffing       `json:"sniffing"`
	Tag      string         `json:"tag"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
	RouteOnly    bool     `json:"routeOnly"`
}

type Log struct {
	Loglevel string `json:"loglevel"`
}

type Outbound struct {
	Protocol       string           `json:"protocol"`
	Settings       OutboundSettings `json:"settings"`
	StreamSettings StreamSettings   `json:"streamSettings"`
	Tag            string           `json:"tag"`
	SendThrough    string           `json:"sendThrough,omitempty"`
}

type OutboundSettings struct {
	Vnext          []Vnext        `json:"vnext"`
	DomainStrategy string         `json:"domainStrategy,omitempty"`
	Redirect       string         `json:"redirect,omitempty"`
	Response       map[string]any `json:"response,omitempty"`
}

type Vnext struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Users   []User `json:"users"`
}

type User struct {
	Encryption string `json:"encryption"`
	Flow       string `json:"flow"`
	ID         string `json:"id"`
}

type StreamSettings struct {
	Network         string           `json:"network,omitempty"`
	Security        string           `json:"security,omitempty"`
	RealitySettings *RealitySettings `json:"realitySettings,omitempty"`
	XhttpSettings   *XhttpSettings   `json:"xhttpSettings,omitempty"`
	XtlsSettings    *XtlsSettings    `json:"xtlsSettings,omitempty"`
}

type RealitySettings struct {
	Network       string `json:"network"`
	Show          bool   `json:"show"`
	Fingerprint   string `json:"fingerprint"`
	AllowInsecure bool   `json:"allowInsecure"`
	PublicKey     string `json:"publicKey"`
	ShortID       string `json:"shortId"`
	SpiderX       string `json:"spiderX"`
	ServerName    string `json:"serverName"`
	Dest            string           `json:"dest,omitempty"`
}

type XhttpSettings struct {
	Path string `json:"path"`
	Host string `json:"host"`
	Mode string `json:"mode"`
}

type XtlsSettings struct {
	DisableSystemRoot bool `json:"disableSystemRoot"`
}

type Policy struct {
	System map[string]bool `json:"system"`
}

type Routing struct {
	DomainMatcher  string           `json:"domainMatcher"`
	DomainStrategy string           `json:"domainStrategy"`
	Rules          []map[string]any `json:"rules"`
}
