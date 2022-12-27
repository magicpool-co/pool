package route53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/types"
)

var (
	ErrZoneNotFound   = fmt.Errorf("zone not found")
	ErrRecordNotFound = fmt.Errorf("record not found")
)

func GetARecordIPByName(client *aws.Client, zoneID, name string) (string, error) {
	if len(name) == 0 {
		return "", ErrRecordNotFound
	} else if name[:len(name)-1] != "." {
		name += "."
	}

	svc := route53.New(client.Session())
	records, err := svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    types.StringPtr(zoneID),
		StartRecordName: types.StringPtr(name),
		StartRecordType: types.StringPtr("A"),
	})
	if err != nil {
		return "", err
	}

	for _, record := range records.ResourceRecordSets {
		if types.StringValue(record.Name) == name {
			for _, ip := range record.ResourceRecords {
				return types.StringValue(ip.Value), nil
			}
		}
	}

	return "", ErrRecordNotFound
}

func GetZoneIDByName(client *aws.Client, name string) (string, error) {
	if len(name) == 0 {
		return "", ErrZoneNotFound
	} else if name[:len(name)-1] != "." {
		name += "."
	}

	svc := route53.New(client.Session())
	zones, err := svc.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName: types.StringPtr(name),
	})
	if err != nil {
		return "", err
	}

	for _, zone := range zones.HostedZones {
		if types.StringValue(zone.Name) == name {
			return strings.ReplaceAll(types.StringValue(zone.Id), "/hostedzone/", ""), nil
		}
	}

	return "", ErrRecordNotFound
}

func UpdateARecords(client *aws.Client, zone string, records map[string]string) error {
	var count int
	changes := make([]*route53.Change, len(records))
	for record, ip := range records {
		changes[count] = &route53.Change{
			Action: types.StringPtr("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: types.StringPtr(record),
				Type: types.StringPtr("A"),
				TTL:  types.Int64Ptr(60),
				ResourceRecords: []*route53.ResourceRecord{
					&route53.ResourceRecord{
						Value: types.StringPtr(ip),
					},
				},
			},
		}
		count++
	}

	svc := route53.New(client.Session())
	_, err := svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		HostedZoneId: types.StringPtr(zone),
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
	})

	return err
}
