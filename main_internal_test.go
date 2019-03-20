package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type wrapper func()

func outputHelper(fn wrapper) (stdout []byte, stderr []byte) {
	outReader, outWriter, _ := os.Pipe()
	errReader, errWriter, _ := os.Pipe()
	originalOut := os.Stdout
	originalErr := os.Stderr
	defer func() {
		os.Stdout = originalOut
		os.Stderr = originalErr
	}()
	os.Stdout = outWriter
	os.Stderr = errWriter
	fn()
	outWriter.Close()
	errWriter.Close()
	stdout, _ = ioutil.ReadAll(outReader)
	stderr, _ = ioutil.ReadAll(errReader)
	return
}

func TestRedirectURL(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost: "example.org",
	}).String()
	expectedURL := "http://test.example.org/path"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithoutTargetHost(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/path", // NewRequest function uses example.com as default host
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		AlwaysHTTPS: true,
	}).String()
	expectedURL := "https://example.com/path"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithoutScheme(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/path", // NewRequest function uses example.com as default host
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost: "example.org",
	}).String()
	expectedURL := "http://example.example.org/path"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithForwardedHost(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"/path", // NewRequest function uses example.com as default host
		nil,
	)
	req.Header.Set("X-Forwarded-Host", "test.example.com")
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://test.example.org/path?via=test.example.com"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithQueryOption(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://test.example.org/path?via=test.example.com"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithAlwaysHTTPSOption(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:  "example.org",
		AlwaysHTTPS: true,
	}).String()
	expectedURL := "https://test.example.org/path"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithQuery(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com/path?q=1",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://test.example.org/path?q=1&via=test.example.com"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithPort(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com:8080/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://test.example.org/path?via=test.example.com%3A8080"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithTLD(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://localhost/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://localhost.example.org/path?via=localhost"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithTLDAndPort(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://localhost:8080/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://localhost.example.org/path?via=localhost%3A8080"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectURLWithTargetHostPort(t *testing.T) {
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com/path",
		nil,
	)
	url := redirectURL(req, &redirectOptions{
		TargetHost:     "example.org:8080",
		HostQueryParam: "via",
	}).String()
	expectedURL := "http://test.example.org:8080/path?via=test.example.com"
	if url != expectedURL {
		t.Errorf("Unexpected redirect URL: %s. Expected: %s", url, expectedURL)
	}
}

func TestRedirectHandlerResponse(t *testing.T) {
	targetHost = "example.org"
	hostQueryParam = "via"
	alwaysHTTPS = false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com?q=1",
		nil,
	)
	outputHelper(func() {
		redirectHandler(rec, req)
	})
	resp := rec.Result()
	if resp.StatusCode != 302 {
		t.Errorf("Unexpected status code: %d. Expected: %d", resp.StatusCode, 302)
	}
	location, err := resp.Location()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	targetURL := "http://test.example.org?q=1&via=test.example.com"
	if location.String() != targetURL {
		t.Errorf("Unexpected location: %s. Expected: %s", location, targetURL)
	}
}

func TestRedirectHandlerResponseWithForwardedProto(t *testing.T) {
	targetHost = "example.org"
	hostQueryParam = "via"
	alwaysHTTPS = false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com",
		nil,
	)
	req.Header.Set("X-Forwarded-Proto", "https")
	outputHelper(func() {
		redirectHandler(rec, req)
	})
	resp := rec.Result()
	location, _ := resp.Location()
	targetURL := "https://test.example.org?via=test.example.com"
	if location.String() != targetURL {
		t.Errorf("Unexpected location: %s. Expected: %s", location, targetURL)
	}
}

func TestRedirectHandlerOutput(t *testing.T) {
	targetHost = "example.org"
	hostQueryParam = "via"
	alwaysHTTPS = false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com?q=1",
		nil,
	)
	req.Header.Set("Referer", "http://example.com/")
	req.Header.Set("User-Agent", "Examplebot/1.0 (+http://example.com)")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "test.example.com")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	timeBefore := time.Now()
	stdout, stderr := outputHelper(func() {
		redirectHandler(rec, req)
	})
	timeAfter := time.Now()
	if string(stderr) != "" {
		t.Errorf("Unexpected stderr: %s", stderr)
	}
	var entry logEntry
	json.Unmarshal(stdout, &entry)
	if entry.Time.Before(timeBefore) {
		t.Errorf("Unexpected 'Time' log: %s", entry.Time)
	}
	if entry.Time.After(timeAfter) {
		t.Errorf("Unexpected 'Time' log: %s", entry.Time)
	}
	if entry.RemoteIP != "192.0.2.1" {
		t.Errorf(
			"Unexpected 'IP' log: %s. Expected: %s",
			entry.RemoteIP,
			"192.0.2.1",
		)
	}
	if entry.Method != "GET" {
		t.Errorf("Unexpected 'Method' log: %s. Expected: %s", entry.Method, "GET")
	}
	if entry.Host != "test.example.com" {
		t.Errorf(
			"Unexpected 'Host' log: %s. Expected: %s",
			entry.Host,
			"test.example.com",
		)
	}
	if entry.RequestURI != "/?q=1" {
		t.Errorf(
			"Unexpected 'RequestURI' log: %s. Expected: %s",
			entry.RequestURI,
			"/?q=1",
		)
	}
	if entry.Referrer != "http://example.com/" {
		t.Errorf(
			"Unexpected 'Referrer' log: %s. Expected: %s",
			entry.Referrer,
			"http://example.com/",
		)
	}
	if entry.UserAgent != "Examplebot/1.0 (+http://example.com)" {
		t.Errorf(
			"Unexpected 'UserAgent' log: %s. Expected: %s",
			entry.UserAgent,
			"Examplebot/1.0 (+http://example.com)",
		)
	}
	if entry.ForwardedFor != "127.0.0.1" {
		t.Errorf(
			"Unexpected 'ForwardedFor' log: %s. Expected: %s",
			entry.ForwardedFor,
			"127.0.0.1",
		)
	}
	if entry.ForwardedHost != "test.example.com" {
		t.Errorf(
			"Unexpected 'ForwardedHost' log: %s. Expected: %s",
			entry.ForwardedHost,
			"test.example.com",
		)
	}
	if entry.ForwardedProto != "https" {
		t.Errorf(
			"Unexpected 'ForwardedProto' log: %s. Expected: %s",
			entry.ForwardedProto,
			"https",
		)
	}
}

func TestRedirectHandlerOutputWithIPv6(t *testing.T) {
	targetHost = "example.org"
	hostQueryParam = "via"
	alwaysHTTPS = false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		"GET",
		"http://test.example.com",
		nil,
	)
	req.RemoteAddr = "[::1]:1234"
	stdout, _ := outputHelper(func() {
		redirectHandler(rec, req)
	})
	var entry logEntry
	json.Unmarshal(stdout, &entry)
	if entry.RemoteIP != "::1" {
		t.Errorf(
			"Unexpected 'IP' log: %s. Expected: %s",
			entry.RemoteIP,
			"192.0.2.1",
		)
	}
}
