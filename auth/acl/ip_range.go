package acl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/bldsoft/gost/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
)

var invalidBsonValue = fmt.Errorf("invalid bson string value")

type IpRange struct {
	ips   []net.IP
	cidrs []*net.IPNet

	mu   *sync.RWMutex
	tree *utils.IPTreeSet
}

func MustIpRangeFromStrings(strs ...string) IpRange {
	ipRange, err := IpRangeFromStrings(strs...)
	if err != nil {
		panic(err)
	}
	return ipRange
}

func IpRangeFromStrings(strs ...string) (res IpRange, err error) {
	for _, s := range strs {
		if strings.Contains(s, "/") {
			_, network, err := net.ParseCIDR(s)
			if err != nil {
				return res, err
			}
			res.cidrs = append(res.cidrs, network)
		} else {
			ip := net.ParseIP(s)
			if ip == nil {
				return res, errors.New("unable to parse IP address")
			}
			res.ips = append(res.ips, ip)
		}
	}
	res.buildTree()
	return res, nil
}

func (r IpRange) Empty() bool {
	return len(r.ips) == 0 && len(r.cidrs) == 0
}

func (r *IpRange) IPs() []net.IP {
	return r.ips
}

func (r *IpRange) CIDRs() []*net.IPNet {
	return r.cidrs
}

func (r *IpRange) SetIPs(ips []net.IP) {
	r.ips = ips
	r.buildTree()
}

func (r *IpRange) SetCIDRs(cidrs []*net.IPNet) {
	r.cidrs = cidrs
	r.buildTree()
}

func (r *IpRange) Strings() []string {
	res := make([]string, 0, len(r.ips)+len(r.cidrs))
	for _, ip := range r.ips {
		res = append(res, ip.String())
	}
	for _, cidr := range r.cidrs {
		res = append(res, cidr.String())
	}
	return res
}

func (r *IpRange) String() string {
	return strings.Join(r.Strings(), ",")
}

func (r *IpRange) buildTree() {
	if r.mu == nil {
		r.mu = &sync.RWMutex{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tree = utils.NewIPTreeSet(r.Strings()...)
}

func (r *IpRange) Contains(ip net.IP) bool {
	if r.mu == nil {
		r.buildTree()
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tree.Match(ip)
}

func (r IpRange) MarshalJSON() ([]byte, error) {
	if r.Empty() { // TODO: remove
		return json.Marshal(nil)
	} //
	return json.Marshal(r.Strings())
}

func (r *IpRange) UnmarshalJSON(data []byte) error {
	var strs []string

	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	ipRange, err := IpRangeFromStrings(strs...)
	if err != nil {
		return err
	}
	*r = ipRange
	return nil
}

func (r IpRange) MarshalBSONValue() (byte, []byte, error) {
	t, data, err := bson.MarshalValue(r.Strings())
	return byte(t), data, err
}

func (r *IpRange) UnmarshalBSONValue(b byte, value []byte) error {
	if value == nil {
		return nil
	}
	t := bson.Type(b)
	if t != bson.TypeArray {
		return fmt.Errorf("invalid bson value type '%s'", t.String())
	}

	arr, _, ok := bsoncore.ReadArray(value)
	if !ok {
		return invalidBsonValue
	}

	values, err := arr.Values()
	if err != nil {
		return invalidBsonValue
	}

	var strs []string
	for _, value := range values {
		strs = append(strs, value.StringValue())
	}

	ipRange, err := IpRangeFromStrings(strs...)
	if err != nil {
		return err
	}
	*r = ipRange
	return nil
}
