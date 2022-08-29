package jwt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
)

const (
	// keys
	RSA_PRIVATE_KEY = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAxEvtIl346kw4VU5vUyjq4Ke1CIRxieiEzHmLTYfLv8mhrrbn
urp77e78QpSrzFREfpxx3DBzI+HYz2E2Fa53LDHa1TZRt2MXjGGue/ZcxtMff9KY
NrcDKSr8bzF5jQV7KcUpCE0pU6fjcsexnZ8oP9aTjIweMpxD/1uCdzT2b2zBCCHW
DmqW02L4mANTXzx2D92yI8z0BeBN7egGWsreyzcqeJoQYn9F8nNwpcORW8/sdET7
qNiwi0i/3sKfAWlo4ghCJvWjPZ7ol0zTeP8S/01fDzcJ+RwYutRwq9tSsa2OX6mu
LxU30GJyTykIqoOsmice5XK7wYvuddDwbzvCZwIDAQABAoIBAA2VT3B8XjggVBmb
FtsDTtWC5WUUfSLU4zOfCYOw7Ol1k2maoIhsVR0O14fn78dw4Vh9jOd2ttD51N3s
1ShE7VwyivkMDDWLdtqw+7j6QEtF2r0mnn/SxDY6EkPDgay/f1lRmlhHtp9iU7vi
k6jblZZtS8KDA6fu5kmqVGnZUWA+YD71reTSwJbxcG7r7pvM5E/pmaRcZy96PJ/B
Gp1I9PDCZWzgx/al/ZzQKrOBvDJ1FwqnbYL0tJ9OWizFECXXHXOMi0uYT0VSyilJ
w80arYKkKuq1o6F4/vqSfdMLWue4qL7FNSKoGndBm/71voGPQg6aAzggvuPTlm2/
vqdgx/ECgYEA7M2FbYJdyfOWhpPhGER5KeAChqraTEK/VnOMz7BYdqeQWs9rA42k
lGJFzVJcU1EgZ2L+SL8xivE35X8PNbgzP41fHoKZR4mMsk4VaiM6Iyg4MVuiLtSD
5Yp99w5gkGkbI4tmq7Qa0Yyb1InVIzoYJ8JL5QQnNK517CLnxWeZpx8CgYEA1DXC
sIx3EnAGvJ9ZonqzCKCl2ZDhS1xPTmlEcYViMWYhaOgdG6dNRaRmV05pPByQY0S7
XxAjsvOYwc+vEbzcyO7wN+ZRBPXx1sS7nKfDjCG6Q0Wy3j3Kz1p6sYbVQDgAEq/M
PptsXWG/GB815SPGQWTAT0RA6f+yXFekjw1PY7kCgYBhbLDvfApAMyKD3cVnKYy7
6LjBVPuZEoEL/WA6dm/+6TOf2ORLQvQqREA5mB/5+0+cmYLKxTaJ1nJLzjmgvVcA
V5aBw/NyFio3lZ6D21ho7Hwp+mxAXhih0JfAlD6wSz3qIskr7V53RiU6jTaOVrFn
ci2tXEcRCpHjg/zdH6F8uwKBgBcRUPyICFmEu/a9C61R0Sxa6ixgR109x5EqeDou
2aGtDGyu7psW8JtlZ4qOB3p1UGy6B/QpePf26uAGh21SLl3ZO1uYOa5kXcmO0SYS
RntxHyI47VyjMuyfVT7+/Sdh7wAZBAa6NmlgOrmQivdBkEeDgDQdo0DMfsLy8/xo
4fxhAoGAGddv4rWkn8oZDQLuQa3apVs+DUZRH81M+FRG38/KMnxMvvpco7PTqQyn
FskadurBfveMGP/eRl2XUIlxKFPzkuT0rSPWARboPhvDnYU/VYZWwAOnm9gLX5n4
INuDIZYuPyjjSLkR5/ttGYt/IVcowScb8cbVIfZQX+sq4qoAohg=
-----END RSA PRIVATE KEY-----`
	RSA_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxEvtIl346kw4VU5vUyjq
4Ke1CIRxieiEzHmLTYfLv8mhrrbnurp77e78QpSrzFREfpxx3DBzI+HYz2E2Fa53
LDHa1TZRt2MXjGGue/ZcxtMff9KYNrcDKSr8bzF5jQV7KcUpCE0pU6fjcsexnZ8o
P9aTjIweMpxD/1uCdzT2b2zBCCHWDmqW02L4mANTXzx2D92yI8z0BeBN7egGWsre
yzcqeJoQYn9F8nNwpcORW8/sdET7qNiwi0i/3sKfAWlo4ghCJvWjPZ7ol0zTeP8S
/01fDzcJ+RwYutRwq9tSsa2OX6muLxU30GJyTykIqoOsmice5XK7wYvuddDwbzvC
ZwIDAQAB
-----END PUBLIC KEY-----`
	ECDSA_PRIVATE_KEY = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIJP7MRLpOmbTfL2EY3yCWhoY0H8HHiipYllK557FJrVboAoGCCqGSM49
AwEHoUQDQgAECvn/657X+odo+dcbAATe+bGMuEl4crTkxT6nk3/0JQYFCD+Ooz2C
Aq8yeag8ni1OaeGqudM+w14iu15fzeJHiw==
-----END EC PRIVATE KEY-----`
	ECDSA_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAECvn/657X+odo+dcbAATe+bGMuEl4
crTkxT6nk3/0JQYFCD+Ooz2CAq8yeag8ni1OaeGqudM+w14iu15fzeJHiw==
-----END PUBLIC KEY-----
`
)

func testJwtConfig(t *testing.T, signerCfg, clientCfg JwtConfig) {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	handler = JwtAuthMiddlewareFromConfig(clientCfg)(handler)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)

	token, _ := jwt.NewBuilder().Expiration(time.Now().Add(1 * time.Minute)).Build()
	signed, _ := SignToken(token, signerCfg)
	r.Header.Add("Authorization", "Bearer "+string(signed))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestJwtConfigPrivateKey(t *testing.T) {
	tests := []struct {
		Alg string
		Key string
	}{
		{"RS256", RSA_PRIVATE_KEY},
		{"ES256", ECDSA_PRIVATE_KEY},
	}
	for _, test := range tests {
		t.Run(test.Alg, func(t *testing.T) {
			cfg := JwtConfig{
				Alg:     test.Alg,
				PemPath: filepath.Join(t.TempDir(), "key"),
			}
			assert.NoError(t, ioutil.WriteFile(cfg.PemPath, []byte(test.Key), 0666))
			assert.NoError(t, cfg.Validate())
			testJwtConfig(t, cfg, cfg)
		})
	}
}

func TestJwtConfigPublicKey(t *testing.T) {
	tests := []struct {
		Alg        string
		PrivateKey string
		PublicKey  string
	}{
		{"RS256", RSA_PRIVATE_KEY, RSA_PUBLIC_KEY},
		{"ES256", ECDSA_PRIVATE_KEY, ECDSA_PUBLIC_KEY},
	}
	dir := t.TempDir()
	for _, test := range tests {
		t.Run(test.Alg, func(t *testing.T) {
			var signerCfg, clientCfg JwtConfig
			for _, cfg := range []struct {
				jwtCfg  *JwtConfig
				pemPath string
				key     string
			}{
				{&signerCfg, filepath.Join(dir, "private"), test.PrivateKey},
				{&clientCfg, filepath.Join(dir, "public"), test.PublicKey},
			} {
				cfg.jwtCfg.Alg = test.Alg
				cfg.jwtCfg.PemPath = cfg.pemPath
				assert.NoError(t, ioutil.WriteFile(cfg.pemPath, []byte(cfg.key), 0666))
				assert.NoError(t, cfg.jwtCfg.Validate())
			}
			testJwtConfig(t, signerCfg, clientCfg)
		})
	}
}
