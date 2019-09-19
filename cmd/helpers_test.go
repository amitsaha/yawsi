package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
)

type mockRoute53Client struct {
	route53iface.Route53API
}

func (m *mockRoute53Client) ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {

	resp := route53.ListResourceRecordSetsOutput{}

	r1 := route53.ResourceRecordSet{}

	mxRecordName := "foo.example.com"
	mxRecordType := "MX"

	mxRecordValues := []string{
		"5 ALT1.ASPMX.L.GOOGLE.COM.",
		"5 ALT2.ASPMX.L.GOOGLE.COM.",
		"10 ASPMX2.GOOGLEMAIL.COM.",
		"1 ASPMX.L.GOOGLE.COM.",
		"10 ASPMX3.GOOGLEMAIL.COM.",
	}

	var mxRecords []*route53.ResourceRecord
	for _, v := range mxRecordValues {
		mxRecords = append(mxRecords, &route53.ResourceRecord{
			Value: &v,
		})
	}

	r1.Name = &mxRecordName
	r1.Type = &mxRecordType
	r1.ResourceRecords = mxRecords

	resourceRecords := []*route53.ResourceRecordSet{&r1}
	resp.ResourceRecordSets = resourceRecords

	return &resp, nil

}
func TestListR53Records(t *testing.T) {

	mockSvc := &mockRoute53Client{}
	rs := ListR53RecordSets(mockSvc, "example.org")
	if len(rs.ResourceRecordSets) != 1 {
		t.Errorf("Expected 1 resource record set, got %v \n", len(rs.ResourceRecordSets))
	}

	expectedMxRecordValues := map[string]int{
		"5 ALT1.ASPMX.L.GOOGLE.COM.": 1,
		"5 ALT2.ASPMX.L.GOOGLE.COM.": 1,
		"10 ASPMX2.GOOGLEMAIL.COM.":  1,
		"1 ASPMX.L.GOOGLE.COM.":      1,
		"10 ASPMX3.GOOGLEMAIL.COM.":  1,
	}

	records := rs.ResourceRecordSets[0].ResourceRecords
	for _, r := range records {
		if _, ok := expectedMxRecordValues[*r.Value]; !ok {
			t.Errorf("Expected record value %v, Not found.", *r.Value)
		}
	}
}

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	cfClient *cloudflare.API

	// server is a test HTTP server used to provide mock API responses
	server *httptest.Server
)

func CloudflareTeardown() {
	server.Close()
}
func CloudflareSetup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	cfClient, _ = cloudflare.New("apikey", "test@example.org")
	cfClient.BaseURL = server.URL
}

func TestCopyRoute53RecordsToCloudflare(t *testing.T) {
	CloudflareSetup()
	defer CloudflareTeardown()

	getZoneHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
			  {
				"id": "023e105f4ecef8ad9ca31a8372d0c353",
				"name": "example.org",
				"development_mode": 7200,
				"original_name_servers": [
				  "ns1.originaldnshost.com",
				  "ns2.originaldnshost.com"
				],
				"original_registrar": "GoDaddy",
				"original_dnshost": "NameCheap",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"modified_on": "2014-01-01T05:20:00.12345Z",
				"activated_on": "2014-01-02T00:01:00.12345Z",
				"owner": {
				  "id": {},
				  "email": {},
				  "type": "user"
				},
				"account": {
				  "id": "01a7362d577a6c3019a474fd6f485823",
				  "name": "Demo Account"
				},
				"permissions": [
				  "#zone:read",
				  "#zone:edit"
				],
				"plan": {
				  "id": "e592fd9519420ba7405e1307bff33214",
				  "name": "Pro Plan",
				  "price": 20,
				  "currency": "USD",
				  "frequency": "monthly",
				  "legacy_id": "pro",
				  "is_subscribed": true,
				  "can_subscribe": true
				},
				"plan_pending": {
				  "id": "e592fd9519420ba7405e1307bff33214",
				  "name": "Pro Plan",
				  "price": 20,
				  "currency": "USD",
				  "frequency": "monthly",
				  "legacy_id": "pro",
				  "is_subscribed": true,
				  "can_subscribe": true
				},
				"status": "active",
				"paused": false,
				"type": "full",
				"name_servers": [
				  "tony.ns.cloudflare.com",
				  "woz.ns.cloudflare.com"
				]
			  }
			]
		  }
		`)
	}

	// test case: No existing DNS records for zone
	getZoneDNSRecordsHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{}`)
	}

	createDNSRecordHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `
		{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "id": "372e67954025e0ba6aaa6d586b9e0b59",
			  "type": "MX",
			  "name": "example.org",
			  "content": "198.51.100.4",
			  "proxiable": true,
			  "proxied": false,
			  "ttl": {},
			  "locked": false,
			  "zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
			  "zone_name": "example.com",
			  "created_on": "2014-01-01T05:20:00.12345Z",
			  "modified_on": "2014-01-01T05:20:00.12345Z",
			  "data": {}
			}
		  }
		`)
	}

	mux.HandleFunc("/zones/?name=example.org", getZoneHandler)
	mux.HandleFunc("/zones/023e105f4ecef8ad9ca31a8372d0c353/dns_records", getZoneDNSRecordsHandler)

}
