package routing

import (
	"net"
	"net/http"
	"net/url"
	"path/filepath"
)

type HostExtractor struct{}

func (e HostExtractor) ExtractValue(r *http.Request) string { return r.Host }

var Host = HostExtractor{}

//=============================================================================

type IpExtractor struct{}

func (e IpExtractor) ExtractValue(r *http.Request) net.IP { return net.ParseIP(r.RemoteAddr) }

var IP = IpExtractor{}

//=============================================================================

type PathExtractor struct{}

func (e PathExtractor) ExtractValue(r *http.Request) string { return r.URL.Path }

var Path = PathExtractor{}

//=============================================================================

type FileNameExtractor struct{}

func (e FileNameExtractor) ExtractValue(r *http.Request) string {
	_, file := filepath.Split(r.URL.Path)
	return file
}

var FileName = FileNameExtractor{}

//=============================================================================

type FileExtExtractor struct{}

func (e FileExtExtractor) ExtractValue(r *http.Request) string { return filepath.Ext(r.URL.Path) }

var FileExt = FileExtExtractor{}

//=============================================================================

type UserAgentExtractor struct{}

func (e UserAgentExtractor) ExtractValue(r *http.Request) string { return r.UserAgent() }

var UserAgent = UserAgentExtractor{}

//=============================================================================

type QueryExtractor struct{}

func (e QueryExtractor) ExtractValue(r *http.Request) url.Values { return r.URL.Query() }

var Query = QueryExtractor{}

//=============================================================================

type HeaderExtractor struct{}

func (e HeaderExtractor) ExtractValue(r *http.Request) http.Header { return r.Header }

var Header = HeaderExtractor{}
