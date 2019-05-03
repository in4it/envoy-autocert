package envoy

import (
	"fmt"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/types"
)

type Listener struct{}

func newListener() *Listener {
	return &Listener{}
}

func (l *Listener) updateListenerWithNewCert(cache *WorkQueueCache, params TLSParams) error {
	var listenerFound bool
	for listenerKey := range cache.listeners {
		ll := cache.listeners[listenerKey].(*api.Listener)
		if ll.Name == "l_"+params.Name+"_tls" {
			listenerFound = true
			logger.Debugf("Matching listener found, updating: %s", ll.Name)
			// add cert and key to tls listener
			ll.FilterChains[0].TlsContext = &auth.DownstreamTlsContext{
				CommonTlsContext: &auth.CommonTlsContext{
					TlsCertificates: []*auth.TlsCertificate{
						{
							CertificateChain: &core.DataSource{
								Specifier: &core.DataSource_InlineString{
									InlineString: params.CertBundle,
								},
							},
							PrivateKey: &core.DataSource{
								Specifier: &core.DataSource_InlineString{
									InlineString: params.PrivateKey,
								},
							},
						},
					},
				},
			}
		}
	}
	if !listenerFound {
		return fmt.Errorf("No tls listener found")
	}
	return nil
}

func (l *Listener) updateListenerWithChallenge(cache *WorkQueueCache, challenge ChallengeParams) error {
	clusterName := challenge.Name
	logger.Debugf("Update listener with challenge for: %s", clusterName)
	newRoute := []route.Route{
		{
			Match: route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
					Path: "/.well-known/acme-challenge/" + challenge.Token,
				},
			},
			Action: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: 200,
					Body: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: challenge.Body,
						},
					},
				},
			},
		},
	}
	for listenerKey := range cache.listeners {
		ll := cache.listeners[listenerKey].(*api.Listener)
		if ll.Name == "l_"+clusterName {
			logger.Debugf("Matching listener found, updating: %s", ll.Name)
			manager, err := l.getListenerHTTPConnectionManager(ll)
			if err != nil {
				return err
			}
			routeSpecifier, err := l.getListenerRouteSpecifier(manager)
			if err != nil {
				return err
			}
			for k, virtualHost := range routeSpecifier.RouteConfig.VirtualHosts {
				if virtualHost.Name == clusterName+"_service" {
					routeSpecifier.RouteConfig.VirtualHosts[k].Routes = append(newRoute, routeSpecifier.RouteConfig.VirtualHosts[k].Routes...)
				}
			}
			manager.RouteSpecifier = routeSpecifier
			pbst, err := types.MarshalAny(&manager)
			if err != nil {
				panic(err)
			}
			ll.FilterChains[0].Filters[0].ConfigType = &listener.Filter_TypedConfig{
				TypedConfig: pbst,
			}
			logger.Debugf("Created new typedConfig: %+v", cache.listeners[listenerKey])
		}
	}
	return nil
}
func (l *Listener) getListenerRouteSpecifier(manager hcm.HttpConnectionManager) (*hcm.HttpConnectionManager_RouteConfig, error) {
	var routeSpecifier *hcm.HttpConnectionManager_RouteConfig
	routeSpecifier = manager.RouteSpecifier.(*hcm.HttpConnectionManager_RouteConfig)
	if len(routeSpecifier.RouteConfig.VirtualHosts) == 0 {
		return routeSpecifier, fmt.Errorf("No virtualhosts found in routeconfig")
	}
	return routeSpecifier, nil
}
func (l *Listener) getListenerHTTPConnectionManager(ll *api.Listener) (hcm.HttpConnectionManager, error) {
	var manager hcm.HttpConnectionManager
	if len(ll.FilterChains) == 0 {
		return manager, fmt.Errorf("No filterchains found in listener %s", ll.Name)
	}
	if len(ll.FilterChains[0].Filters) == 0 {
		return manager, fmt.Errorf("No filters found in listener %s", ll.Name)
	}
	typedConfig := (ll.FilterChains[0].Filters[0].ConfigType).(*listener.Filter_TypedConfig)
	err := types.UnmarshalAny(typedConfig.TypedConfig, &manager)
	if err != nil {
		return manager, err
	}
	return manager, nil
}
func (l *Listener) getVirtualHost(hostname, targetHostname, targetPrefix, clusterName, virtualHostName string) route.VirtualHost {
	return route.VirtualHost{
		Name:    virtualHostName,
		Domains: []string{hostname},

		Routes: []route.Route{{
			Match: route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: targetPrefix,
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					HostRewriteSpecifier: &route.RouteAction_HostRewrite{
						HostRewrite: targetHostname,
					},
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}}}
}
func (l *Listener) updateListener(cache *WorkQueueCache, params ListenerParams, paramsTLS TLSParams) error {
	var listenerKey = -1

	_, targetPrefix, virtualHostname, _, listenerName, _ := l.getListenerAttributes(params, paramsTLS)

	logger.Infof("Updating listener " + listenerName)

	for k, listener := range cache.listeners {
		if (listener.(*api.Listener)).Name == listenerName {
			listenerKey = k
		}
	}
	if listenerKey == -1 {
		return fmt.Errorf("No matching listener found")
	}
	// update listener
	ll := cache.listeners[listenerKey].(*api.Listener)
	manager, err := l.getListenerHTTPConnectionManager(ll)
	if err != nil {
		return err
	}
	routeSpecifier, err := l.getListenerRouteSpecifier(manager)
	if err != nil {
		return err
	}

	// create new virtualhost
	v := l.getVirtualHost(params.Conditions.Hostname, params.TargetHostname, targetPrefix, params.Name, virtualHostname)

	// append new virtualhost
	routeSpecifier.RouteConfig.VirtualHosts = append(routeSpecifier.RouteConfig.VirtualHosts, v)

	manager.RouteSpecifier = routeSpecifier
	pbst, err := types.MarshalAny(&manager)
	if err != nil {
		panic(err)
	}
	ll.FilterChains[0].Filters[0].ConfigType = &listener.Filter_TypedConfig{
		TypedConfig: pbst,
	}
	logger.Debugf("Updated listener with new Virtualhost")

	return nil
}
func (l *Listener) getListenerAttributes(params ListenerParams, paramsTLS TLSParams) (bool, string, string, string, string, uint32) {
	var (
		tls             bool
		listenerName    string
		targetPrefix    = "/"
		virtualHostName string
		routeConfigName string
		listenerPort    uint32
	)

	if paramsTLS.CertBundle != "" {
		tls = true
	}

	if params.Conditions.Prefix != "" && params.Conditions.Prefix != "/" {
		targetPrefix = params.Conditions.Prefix
	}

	if params.Conditions.Hostname == "" {
		virtualHostName = params.Name + "_service" + "_wildcard"
		routeConfigName = params.Name + "_route" + "_wildcard"
	} else {
		virtualHostName = params.Name + "_service" + "_" + params.Conditions.Hostname
		routeConfigName = params.Name + "_route" + "_" + params.Conditions.Hostname
	}

	if tls {
		listenerPort = 10001
		listenerName = "l_" + params.Name + "_tls"
		virtualHostName = virtualHostName + "_tls"
		routeConfigName = routeConfigName + "_tls"
	} else {
		listenerPort = 10000
		listenerName = "l_" + params.Name
	}
	return tls, targetPrefix, virtualHostName, routeConfigName, listenerName, listenerPort
}
func (l *Listener) createListener(params ListenerParams, paramsTLS TLSParams) *api.Listener {
	var err error

	tls, targetPrefix, virtualHostName, routeConfigName, listenerName, listenerPort := l.getListenerAttributes(params, paramsTLS)

	logger.Infof("Creating listener " + listenerName)

	v := l.getVirtualHost(params.Conditions.Hostname, params.TargetHostname, targetPrefix, params.Name, virtualHostName)

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &api.RouteConfiguration{
				Name:         routeConfigName,
				VirtualHosts: []route.VirtualHost{v},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
	}

	pbst, err := types.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	listener := &api.Listener{
		Name: listenerName,
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: listenerPort,
					},
				},
			},
		},
		FilterChains: []listener.FilterChain{{
			Filters: []listener.Filter{{
				Name: util.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
	if tls {
		// add cert and key to tls listener
		listener.FilterChains[0].TlsContext = &auth.DownstreamTlsContext{
			CommonTlsContext: &auth.CommonTlsContext{
				TlsCertificates: []*auth.TlsCertificate{
					{
						CertificateChain: &core.DataSource{
							Specifier: &core.DataSource_InlineString{
								InlineString: paramsTLS.CertBundle,
							},
						},
						PrivateKey: &core.DataSource{
							Specifier: &core.DataSource_InlineString{
								InlineString: paramsTLS.PrivateKey,
							},
						},
					},
				},
			},
		}
	}
	return listener
}