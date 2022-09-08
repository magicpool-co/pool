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

	route53Svc := route53.New(client.Session())
	records, err := route53Svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
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

	route53Svc := route53.New(client.Session())
	zones, err := route53Svc.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
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
