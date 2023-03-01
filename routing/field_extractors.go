package routing

import (
	"net"
	"net/http"
	"path/filepath"
	"strings"
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

type QueryExtractor struct {
	Name string `label:"Query name"`
}

func (e QueryExtractor) ExtractValue(r *http.Request) *string {
	values := r.URL.Query()[e.Name]
	if len(values) == 0 {
		return nil
	}
	res := strings.Join(values, ",")
	return &res
}

func Query(name string) QueryExtractor {
	return QueryExtractor{Name: name}
}

//=============================================================================

type HeaderExtractor struct {
	Name string `label:"Header name"`
}

func (e HeaderExtractor) ExtractValue(r *http.Request) *string {
	if h := r.Header.Get(e.Name); len(h) == 0 {
		return &h
	}
	return nil
}

func Header(name string) HeaderExtractor {
	return HeaderExtractor{Name: name}
}
