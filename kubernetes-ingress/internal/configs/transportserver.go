package configs

import (
	"fmt"
	"strings"

	api_v1 "k8s.io/api/core/v1"

	"github.com/nginxinc/kubernetes-ingress/internal/configs/version2"
	"github.com/nginxinc/kubernetes-ingress/internal/k8s/secrets"
	conf_v1alpha1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1alpha1"
)

const nginxNonExistingUnixSocket = "unix:/var/lib/nginx/non-existing-unix-socket.sock"

// TransportServerEx holds a TransportServer along with the resources referenced by it.
type TransportServerEx struct {
	ListenerPort     int
	TransportServer  *conf_v1alpha1.TransportServer
	Endpoints        map[string][]string
	PodsByIP         map[string]string
	ExternalNameSvcs map[string]bool
	DisableIPV6      bool
	SecretRefs       map[string]*secrets.SecretReference
}

func (tsEx *TransportServerEx) String() string {
	if tsEx == nil {
		return "<nil>"
	}
	if tsEx.TransportServer == nil {
		return "TransportServerEx has no TransportServer"
	}
	return fmt.Sprintf("%s/%s", tsEx.TransportServer.Namespace, tsEx.TransportServer.Name)
}

func newUpstreamNamerForTransportServer(transportServer *conf_v1alpha1.TransportServer) *upstreamNamer {
	return &upstreamNamer{
		prefix: fmt.Sprintf("ts_%s_%s", transportServer.Namespace, transportServer.Name),
	}
}

// generateTransportServerConfig generates a full configuration for a TransportServer.
func generateTransportServerConfig(transportServerEx *TransportServerEx, listenerPort int, isPlus bool, isResolverConfigured bool) (*version2.TransportServerConfig, Warnings) {
	warnings := newWarnings()

	upstreamNamer := newUpstreamNamerForTransportServer(transportServerEx.TransportServer)

	upstreams, w := generateStreamUpstreams(transportServerEx, upstreamNamer, isPlus, isResolverConfigured)
	warnings.Add(w)

	healthCheck, match := generateTransportServerHealthCheck(transportServerEx.TransportServer.Spec.Action.Pass,
		upstreamNamer.GetNameForUpstream(transportServerEx.TransportServer.Spec.Action.Pass),
		transportServerEx.TransportServer.Spec.Upstreams)

	sslConfig, w := generateSSLConfig(transportServerEx.TransportServer, transportServerEx.TransportServer.Spec.TLS, transportServerEx.TransportServer.Namespace, transportServerEx.SecretRefs)
	warnings.Add(w)

	var proxyRequests, proxyResponses *int
	var connectTimeout, nextUpstreamTimeout string
	var nextUpstream bool
	var nextUpstreamTries int
	if transportServerEx.TransportServer.Spec.UpstreamParameters != nil {
		proxyRequests = transportServerEx.TransportServer.Spec.UpstreamParameters.UDPRequests
		proxyResponses = transportServerEx.TransportServer.Spec.UpstreamParameters.UDPResponses

		nextUpstream = transportServerEx.TransportServer.Spec.UpstreamParameters.NextUpstream
		if nextUpstream {
			nextUpstreamTries = transportServerEx.TransportServer.Spec.UpstreamParameters.NextUpstreamTries
			nextUpstreamTimeout = transportServerEx.TransportServer.Spec.UpstreamParameters.NextUpstreamTimeout
		}

		connectTimeout = transportServerEx.TransportServer.Spec.UpstreamParameters.ConnectTimeout
	}

	var proxyTimeout string
	if transportServerEx.TransportServer.Spec.SessionParameters != nil {
		proxyTimeout = transportServerEx.TransportServer.Spec.SessionParameters.Timeout
	}

	serverSnippets := generateSnippets(true, transportServerEx.TransportServer.Spec.ServerSnippets, []string{})

	streamSnippets := generateSnippets(true, transportServerEx.TransportServer.Spec.StreamSnippets, []string{})

	statusZone := transportServerEx.TransportServer.Spec.Listener.Name
	if transportServerEx.TransportServer.Spec.Listener.Name == conf_v1alpha1.TLSPassthroughListenerName {
		statusZone = transportServerEx.TransportServer.Spec.Host
	}

	tsConfig := &version2.TransportServerConfig{
		Server: version2.StreamServer{
			TLSPassthrough:           transportServerEx.TransportServer.Spec.Listener.Name == conf_v1alpha1.TLSPassthroughListenerName,
			UnixSocket:               generateUnixSocket(transportServerEx),
			Port:                     listenerPort,
			UDP:                      transportServerEx.TransportServer.Spec.Listener.Protocol == "UDP",
			StatusZone:               statusZone,
			ProxyRequests:            proxyRequests,
			ProxyResponses:           proxyResponses,
			ProxyPass:                upstreamNamer.GetNameForUpstream(transportServerEx.TransportServer.Spec.Action.Pass),
			Name:                     transportServerEx.TransportServer.Name,
			Namespace:                transportServerEx.TransportServer.Namespace,
			ProxyConnectTimeout:      generateTimeWithDefault(connectTimeout, "60s"),
			ProxyTimeout:             generateTimeWithDefault(proxyTimeout, "10m"),
			ProxyNextUpstream:        nextUpstream,
			ProxyNextUpstreamTimeout: generateTimeWithDefault(nextUpstreamTimeout, "0s"),
			ProxyNextUpstreamTries:   nextUpstreamTries,
			HealthCheck:              healthCheck,
			ServerSnippets:           serverSnippets,
			DisableIPV6:              transportServerEx.DisableIPV6,
			SSL:                      sslConfig,
		},
		Match:          match,
		Upstreams:      upstreams,
		StreamSnippets: streamSnippets,
	}
	return tsConfig, warnings
}

func generateUnixSocket(transportServerEx *TransportServerEx) string {
	if transportServerEx.TransportServer.Spec.Listener.Name == conf_v1alpha1.TLSPassthroughListenerName {
		return fmt.Sprintf("unix:/var/lib/nginx/passthrough-%s_%s.sock", transportServerEx.TransportServer.Namespace, transportServerEx.TransportServer.Name)
	}
	return ""
}

func generateSSLConfig(ts *conf_v1alpha1.TransportServer, tls *conf_v1alpha1.TLS, namespace string, secretRefs map[string]*secrets.SecretReference) (*version2.StreamSSL, Warnings) {
	if tls == nil {
		return &version2.StreamSSL{Enabled: false}, nil
	}

	warnings := newWarnings()
	sslEnabled := true

	secretRef := secretRefs[fmt.Sprintf("%s/%s", namespace, tls.Secret)]
	var secretType api_v1.SecretType
	if secretRef.Secret != nil {
		secretType = secretRef.Secret.Type
	}
	name := secretRef.Path
	if secretType != "" && secretType != api_v1.SecretTypeTLS {
		errMsg := fmt.Sprintf("TLS secret %s is of a wrong type '%s', must be '%s'. SSL termination will not be enabled for this server.", tls.Secret, secretType, api_v1.SecretTypeTLS)
		warnings.AddWarning(ts, errMsg)
		sslEnabled = false
	} else if secretRef.Error != nil {
		errMsg := fmt.Sprintf("TLS secret %s is invalid: %v. SSL termination will not be enabled for this server.", tls.Secret, secretRef.Error)
		warnings.AddWarning(ts, errMsg)
		sslEnabled = false
	}

	ssl := version2.StreamSSL{
		Enabled:        sslEnabled,
		Certificate:    name,
		CertificateKey: name,
	}

	return &ssl, warnings
}

func generateStreamUpstreams(transportServerEx *TransportServerEx, upstreamNamer *upstreamNamer, isPlus bool, isResolverConfigured bool) ([]version2.StreamUpstream, Warnings) {
	warnings := newWarnings()
	var upstreams []version2.StreamUpstream

	for _, u := range transportServerEx.TransportServer.Spec.Upstreams {
		// subselector is not supported yet in TransportServer upstreams. That's why we pass "nil" here
		endpointsKey := GenerateEndpointsKey(transportServerEx.TransportServer.Namespace, u.Service, nil, uint16(u.Port))
		externalNameSvcKey := GenerateExternalNameSvcKey(transportServerEx.TransportServer.Namespace, u.Service)
		endpoints := transportServerEx.Endpoints[endpointsKey]

		_, isExternalNameSvc := transportServerEx.ExternalNameSvcs[externalNameSvcKey]
		if isExternalNameSvc && !isResolverConfigured {
			msgFmt := "Type ExternalName service %v in upstream %v will be ignored. To use ExternalName services, a resolver must be configured in the ConfigMap"
			warnings.AddWarningf(transportServerEx.TransportServer, msgFmt, u.Service, u.Name)
			endpoints = []string{}
		}

		ups := generateStreamUpstream(u, upstreamNamer, endpoints, isPlus)
		ups.Resolve = isExternalNameSvc
		ups.UpstreamLabels.Service = u.Service
		ups.UpstreamLabels.ResourceType = "transportserver"
		ups.UpstreamLabels.ResourceName = transportServerEx.TransportServer.Name
		ups.UpstreamLabels.ResourceNamespace = transportServerEx.TransportServer.Namespace

		upstreams = append(upstreams, ups)
	}
	return upstreams, warnings
}

func generateTransportServerHealthCheck(upstreamName string, generatedUpstreamName string, upstreams []conf_v1alpha1.Upstream) (*version2.StreamHealthCheck, *version2.Match) {
	var hc *version2.StreamHealthCheck
	var match *version2.Match

	for _, u := range upstreams {
		if u.Name == upstreamName {
			if u.HealthCheck == nil || !u.HealthCheck.Enabled {
				return nil, nil
			}
			hc = generateTransportServerHealthCheckWithDefaults()

			hc.Enabled = u.HealthCheck.Enabled
			hc.Interval = generateTimeWithDefault(u.HealthCheck.Interval, hc.Interval)
			hc.Jitter = generateTimeWithDefault(u.HealthCheck.Jitter, hc.Jitter)
			hc.Timeout = generateTimeWithDefault(u.HealthCheck.Timeout, hc.Timeout)
			hc.Port = u.HealthCheck.Port

			if u.HealthCheck.Fails > 0 {
				hc.Fails = u.HealthCheck.Fails
			}

			if u.HealthCheck.Passes > 0 {
				hc.Passes = u.HealthCheck.Passes
			}

			if u.HealthCheck.Match != nil {
				name := "match_" + generatedUpstreamName
				match = generateHealthCheckMatch(u.HealthCheck.Match, name)
				hc.Match = name
			}

			break
		}
	}
	return hc, match
}

func generateTransportServerHealthCheckWithDefaults() *version2.StreamHealthCheck {
	return &version2.StreamHealthCheck{
		Enabled:  false,
		Timeout:  "5s",
		Jitter:   "0s",
		Interval: "5s",
		Passes:   1,
		Fails:    1,
		Match:    "",
	}
}

func generateHealthCheckMatch(match *conf_v1alpha1.Match, name string) *version2.Match {
	var modifier string
	var expect string

	if strings.HasPrefix(match.Expect, "~*") {
		modifier = "~*"
		expect = strings.TrimPrefix(match.Expect, "~*")
	} else if strings.HasPrefix(match.Expect, "~") {
		modifier = "~"
		expect = strings.TrimPrefix(match.Expect, "~")
	} else {
		expect = match.Expect
	}

	return &version2.Match{
		Name:                name,
		Send:                match.Send,
		ExpectRegexModifier: modifier,
		Expect:              expect,
	}
}

func generateStreamUpstream(upstream conf_v1alpha1.Upstream, upstreamNamer *upstreamNamer, endpoints []string, isPlus bool) version2.StreamUpstream {
	var upsServers []version2.StreamUpstreamServer

	name := upstreamNamer.GetNameForUpstream(upstream.Name)
	maxFails := generateIntFromPointer(upstream.MaxFails, 1)
	maxConns := generateIntFromPointer(upstream.MaxConns, 0)
	failTimeout := generateTimeWithDefault(upstream.FailTimeout, "10s")

	for _, e := range endpoints {
		s := version2.StreamUpstreamServer{
			Address:        e,
			MaxFails:       maxFails,
			FailTimeout:    failTimeout,
			MaxConnections: maxConns,
		}

		upsServers = append(upsServers, s)
	}

	if !isPlus && len(endpoints) == 0 {
		upsServers = append(upsServers, version2.StreamUpstreamServer{
			Address:     nginxNonExistingUnixSocket,
			MaxFails:    maxFails,
			FailTimeout: failTimeout,
		})
	}

	return version2.StreamUpstream{
		Name:                name,
		Servers:             upsServers,
		LoadBalancingMethod: generateLoadBalancingMethod(upstream.LoadBalancingMethod),
	}
}

func generateLoadBalancingMethod(method string) string {
	if method == "" {
		// By default, if unspecified, Nginx uses the 'round_robin' load balancing method.
		// We override this default which suits the Ingress Controller better.
		return "random two least_conn"
	}
	if method == "round_robin" {
		// By default, Nginx uses round robin. We select this method by not specifying any method.
		return ""
	}
	return method
}
