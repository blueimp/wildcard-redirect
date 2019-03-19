package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	listenAddress  string
	targetHost     string
	hostQueryParam string
	alwaysHTTPS    bool
)

type logEntry struct {
	Time           time.Time
	RemoteIP       string
	Method         string
	Host           string
	RequestURI     string
	Referrer       string
	UserAgent      string
	ForwardedFor   string
	ForwardedProto string
}

func (req *logEntry) Log() {
	b, _ := json.Marshal(req)
	fmt.Fprintln(os.Stdout, string(b))
}

type redirectOptions struct {
	TargetHost     string
	HostQueryParam string
	AlwaysHTTPS    bool
}

func redirectURL(req *http.Request, options *redirectOptions) *url.URL {
	// Copy the request URL struct:
	targetURL := new(url.URL)
	*targetURL = *req.URL
	// Remove any existing user:password authentication:
	targetURL.User = nil
	if options.AlwaysHTTPS {
		// Always redirect to HTTPS:
		targetURL.Scheme = "https"
	} else {
		targetURL.Scheme = "http"
	}
	if len(req.Host) != 0 && len(options.TargetHost) != 0 {
		if options.HostQueryParam != "" {
			query := targetURL.Query()
			query.Set(options.HostQueryParam, req.Host)
			targetURL.RawQuery = query.Encode()
		}
		part := strings.Split(req.Host, ".")[0]
		if strings.IndexByte(part, ':') >= 0 {
			part, _, _ = net.SplitHostPort(part)
		}
		targetURL.Host = part + "." + options.TargetHost
	} else {
		targetURL.Host = req.Host
	}
	return targetURL
}

func redirectHandler(resp http.ResponseWriter, req *http.Request) {
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	forwardedProto := req.Header.Get("X-Forwarded-Proto")
	entry := &logEntry{
		Time:           time.Now().UTC(),
		RemoteIP:       ip,
		Method:         req.Method,
		Host:           req.Host,
		RequestURI:     req.URL.RequestURI(),
		Referrer:       req.Header.Get("Referer"),
		UserAgent:      req.Header.Get("User-Agent"),
		ForwardedFor:   req.Header.Get("X-Forwarded-For"),
		ForwardedProto: forwardedProto,
	}
	defer entry.Log()
	options := &redirectOptions{
		TargetHost:     targetHost,
		HostQueryParam: hostQueryParam,
		AlwaysHTTPS:    alwaysHTTPS || (forwardedProto == "https"),
	}
	http.Redirect(
		resp,
		req,
		redirectURL(req, options).String(),
		http.StatusFound,
	)
}

func main() {
	flag.StringVar(&listenAddress, "a", ":8080", "TCP listen address")
	flag.StringVar(&hostQueryParam, "q", "via", "Original host query parameter")
	flag.BoolVar(&alwaysHTTPS, "s", false, "Always redirect using HTTPS")
	flag.Parse()
	targetHost = flag.Arg(0)
	err := http.ListenAndServe(listenAddress, http.HandlerFunc(redirectHandler))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
