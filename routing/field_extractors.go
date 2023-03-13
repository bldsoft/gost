package routing

import (
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bldsoft/gost/utils"
)

type Incoming struct{}

func (Incoming) IsIncoming() bool { return true }

type Outgoing struct{}

func (Outgoing) IsIncoming() bool { return false }

//=============================================================================

type HostExtractor struct{ Incoming }

func (e HostExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) string { return r.Host }

var Host = HostExtractor{}

//=============================================================================

type IpExtractor struct{ Incoming }

func (e IpExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) net.IP {
	return net.ParseIP(r.RemoteAddr)
}

var IP = IpExtractor{}

//=============================================================================

type PathExtractor struct{ Incoming }

func (e PathExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) string { return r.URL.Path }

var Path = PathExtractor{}

//=============================================================================

type FileNameExtractor struct{ Incoming }

func (e FileNameExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) string {
	_, file := filepath.Split(r.URL.Path)
	return file
}

var FileName = FileNameExtractor{}

//=============================================================================

type FileExtExtractor struct{ Incoming }

func (e FileExtExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) string {
	return filepath.Ext(r.URL.Path)
}

var FileExt = FileExtExtractor{}

//=============================================================================

type UserAgentExtractor struct{ Incoming }

func (e UserAgentExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) string {
	return r.UserAgent()
}

var UserAgent = UserAgentExtractor{}

//=============================================================================

type QueryExtractor struct {
	Incoming
	Name string `label:"Query name"`
}

func (e QueryExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) *string {
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
	Incoming
	Name string `label:"Header name"`
}

func (e HeaderExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) *string {
	if h := r.Header.Get(e.Name); len(h) == 0 {
		return &h
	}
	return nil
}

func Header(name string) HeaderExtractor {
	return HeaderExtractor{Name: name}
}

//=============================================================================

type ResponseCodeExtractor struct{ Outgoing }

func (e ResponseCodeExtractor) ExtractValue(w http.ResponseWriter, r *http.Request) int {
	return w.(*utils.ResponseWriter).Code
}

var StatusCode = ResponseCodeExtractor{}
