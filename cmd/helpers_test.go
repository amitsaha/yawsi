package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
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
	// Setup Test
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
