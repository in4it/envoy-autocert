package envoy

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/juju/loggo"
)

func TestDomainAlreadyExists(t *testing.T) {

	l := newListener()
	var cache WorkQueueCache
	params1 := ListenerParams{
		Name:           "test_1",
		Protocol:       "http",
		TargetHostname: "www.test.inv",
		Conditions: Conditions{
			Hostname: "hostname1.example.com",
			Prefix:   "/test1",
		},
	}
	params2 := ListenerParams{
		Name:           "test_2",
		Protocol:       "http",
		TargetHostname: "www.test.inv",
		Conditions: Conditions{
			Hostname: "hostname1.example.com",
			Prefix:   "/test2",
		},
	}
	paramsTLS1 := TLSParams{}
	listener := l.createListener(params1, paramsTLS1)
	cache.listeners = append(cache.listeners, listener)
	err := l.updateListener(&cache, params2, paramsTLS1)

	if len(cache.listeners) == 0 {
		t.Errorf("Listener is empty (got %d)", len(cache.listeners))
		return
	}
	cachedListener := cache.listeners[0].(*api.Listener)
	if cachedListener.Name != "l_http" {
		t.Errorf("Expected l_http (got %s)", cachedListener.Name)
		return
	}

	manager, err := l.getListenerHTTPConnectionManager(cachedListener)
	routeSpecifier, err := l.getListenerRouteSpecifier(manager)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	if len(routeSpecifier.RouteConfig.VirtualHosts) == 0 {
		t.Errorf("Should have more than 0 virtualhosts")
		return
	}
	if len(routeSpecifier.RouteConfig.VirtualHosts[0].Domains) != 1 {
		t.Errorf("Should have 1 domain")
		return
	}
	if routeSpecifier.RouteConfig.VirtualHosts[0].Domains[0] != "hostname1.example.com" {
		t.Errorf("Only domain in virtualhost should be hostname1.example.com")
		return
	}
}

func TestUpdateListener(t *testing.T) {
	// set debug loglevel
	logger.SetLogLevel(loggo.DEBUG)
	l := newListener()
	var cache WorkQueueCache
	params1 := ListenerParams{
		Name:           "test_1",
		Protocol:       "http",
		TargetHostname: "www.test.inv",
		Conditions: Conditions{
			Hostname: "hostname1.example.com",
			Prefix:   "/test1",
		},
	}
	params2 := ListenerParams{
		Name:           "test_2",
		Protocol:       "http",
		TargetHostname: "www.test.inv",
		Conditions: Conditions{
			Hostname: "hostname2.example.com",
			Prefix:   "/test2",
			Methods:  []string{"GET", "POST"},
		},
	}
	params3 := ListenerParams{
		Name:           "test_3",
		Protocol:       "http",
		TargetHostname: "www.test.inv",
		Conditions: Conditions{
			Hostname: "hostname2.example.com",
			Prefix:   "/test3",
		},
		Auth: Auth{
			JwtProvider: "testJwt",
			Issuer:      "http://issuer.example.com",
			Forward:     true,
			RemoteJwks:  "https://remotejwks.example.com",
		},
	}
	params4 := ListenerParams{
		Name:           "test_4",
		Protocol:       "http",
		TargetHostname: "www.test-tls.inv",
		Conditions: Conditions{
			Hostname: "hostname4.example.com",
			Prefix:   "/test4",
		},
		Auth: Auth{
			JwtProvider: "testJwt",
			Issuer:      "http://issuer.example.com",
			Forward:     true,
			RemoteJwks:  "https://remotejwks.example.com",
		},
	}
	paramsTLS1 := TLSParams{}
	paramsTLS4 := TLSParams{
		Name:       "www.test-tls.inv",
		CertBundle: "certbundle",
		PrivateKey: "privateKey",
	}
	listener := l.createListener(params1, paramsTLS1)
	cache.listeners = append(cache.listeners, listener)

	// validate domain 1
	if err := validateDomain(cache.listeners, params1); err != nil {
		t.Errorf("Validation failed: %s", err)
		return
	}

	// update listener with domain 2

	if err := l.updateListener(&cache, params2, paramsTLS1); err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	// validate domain 1 and 2
	if err := validateDomain(cache.listeners, params1); err != nil {
		t.Errorf("Validation failed: %s", err)
		return
	}
	if err := validateDomain(cache.listeners, params2); err != nil {
		t.Errorf("Validation failed: %s", err)
		return
	}

	// add domain 3
	if err := l.updateListener(&cache, params3, paramsTLS1); err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	// validate domain 3
	if err := validateDomain(cache.listeners, params3); err != nil {
		t.Errorf("Validation failed: %s", err)
		return
	}

	// add domain 4 (TLS)
	TLSListener := l.createListener(params4, paramsTLS4)
	cache.listeners = append(cache.listeners, TLSListener)

	// validate domain 4 (TLS)
	if err := validateDomainTLS(cache.listeners, params4, paramsTLS4); err != nil {
		t.Errorf("Validation failed: %s", err)
		return
	}
}
func validateDomainTLS(listeners []cache.Resource, params ListenerParams, tlsParams TLSParams) error {
	//l := newListener()
	if len(listeners) == 0 {
		return fmt.Errorf("Listener is empty (got %d)", len(listeners))
	}
	cachedListener := listeners[1].(*api.Listener)
	if cachedListener.Name != "l_tls" {
		return fmt.Errorf("Expected l_tls (got %s)", cachedListener.Name)
	}

	return nil
}

func validateDomain(listeners []cache.Resource, params ListenerParams) error {
	l := newListener()
	if len(listeners) == 0 {
		return fmt.Errorf("Listener is empty (got %d)", len(listeners))
	}
	cachedListener := listeners[0].(*api.Listener)
	if cachedListener.Name != "l_http" {
		return fmt.Errorf("Expected l_http (got %s)", cachedListener.Name)
	}

	manager, err := l.getListenerHTTPConnectionManager(cachedListener)
	routeSpecifier, err := l.getListenerRouteSpecifier(manager)
	if err != nil {
		return fmt.Errorf("Error: %s", err)
	}

	domainFound := false
	prefixFound := false
	methodsFound := false

	if params.Conditions.Hostname == "" {
		params.Conditions.Hostname = "*"
	}
	if params.Conditions.Prefix == "/" {
		params.Conditions.Prefix = "/"
	}

	for _, virtualhost := range routeSpecifier.RouteConfig.VirtualHosts {
		for _, domain := range virtualhost.Domains {
			if domain == params.Conditions.Hostname {
				domainFound = true
				for _, r := range virtualhost.Routes {
					if r.Match.PathSpecifier.(*route.RouteMatch_Prefix).Prefix == params.Conditions.Prefix {
						prefixFound = true
					}
					if len(params.Conditions.Methods) > 0 {
						methodsInHeader := []string{}
						for _, v := range r.Match.Headers {
							if v.Name == ":method" {
								methodsInHeader = append(methodsInHeader, v.GetExactMatch())
							}
						}
						sort.Strings(methodsInHeader)
						sort.Strings(params.Conditions.Methods)
						if testEqualityString(params.Conditions.Methods, methodsInHeader) {
							methodsFound = true
						}
					}

				}
			}
		}
	}

	if domainFound != true {
		return fmt.Errorf("Domain not found: %s", params.Conditions.Hostname)
	}
	logger.Debugf("Domain found: %s", params.Conditions.Hostname)

	if prefixFound != true {
		return fmt.Errorf("Prefix not found: %s", params.Conditions.Prefix)
	}
	logger.Debugf("Prefix found: %s", params.Conditions.Prefix)

	if len(params.Conditions.Methods) > 0 && !methodsFound {
		return fmt.Errorf("Methods not found: %s", strings.Join(params.Conditions.Methods, ","))

	}
	if len(params.Conditions.Methods) > 0 {
		logger.Debugf("Methods found: %s", strings.Join(params.Conditions.Methods, ","))
	}

	return nil
}
func testEqualityString(a, b []string) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
