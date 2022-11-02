package acl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var invalidBsonValue = fmt.Errorf("invalid bson string value")

type IpRange struct {
	Ip   []net.IP
	Cidr []*net.IPNet
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
			res.Cidr = append(res.Cidr, network)
		} else {
			ip := net.ParseIP(s)
			if ip == nil {
				return res, errors.New("unable to parse IP address")
			}
			res.Ip = append(res.Ip, ip)
		}
	}
	return res, nil
}

func (r IpRange) Empty() bool {
	return len(r.Ip) == 0 && len(r.Cidr) == 0
}

func (r IpRange) isInIPs(client net.IP, ips []net.IP) bool {
	for _, ip := range ips {
		if client.Equal(ip) {
			return true
		}
	}
	return false
}

func (r *IpRange) Strings() []string {
	res := make([]string, 0, len(r.Ip)+len(r.Cidr))
	for _, ip := range r.Ip {
		res = append(res, ip.String())
	}
	for _, cidr := range r.Cidr {
		res = append(res, cidr.String())
	}
	return res
}

func (r *IpRange) String() string {
	return strings.Join(r.Strings(), ",")
}

func (r IpRange) isInSubnets(ip net.IP, subs []*net.IPNet) bool {
	for _, subnet := range subs {
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func (r IpRange) Contains(ip net.IP) bool {
	return r.isInSubnets(ip, r.Cidr) || r.isInIPs(ip, r.Ip)
}

func (r IpRange) MarshalJSON() ([]byte, error) {
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

func (r IpRange) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(r.Strings())
}

func (r *IpRange) UnmarshalBSONValue(t bsontype.Type, value []byte) error {
	if value == nil {
		return nil
	}
	if t != bsontype.Array {
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
