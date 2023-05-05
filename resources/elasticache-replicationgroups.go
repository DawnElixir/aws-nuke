package resources

import (
	"fmt"
	
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/rebuy-de/aws-nuke/v2/pkg/types"
)

type ElasticacheReplicationGroup struct {
	svc     *elasticache.ElastiCache
	groupID *string
	tags      []*elasticache.Tag
}

func init() {
	register("ElasticacheReplicationGroup", ListElasticacheReplicationGroups)
}

func ListElasticacheReplicationGroups(sess *session.Session) ([]Resource, error) {
	svc := elasticache.New(sess)
        // Lookup current account ID
	stsSvc := sts.New(sess)
	callerID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountID := callerID.Account
	region := svc.Config.Region
	var resources []Resource

	params := &elasticache.DescribeReplicationGroupsInput{MaxRecords: aws.Int64(100)}

	for {
		resp, err := svc.DescribeReplicationGroups(params)
		if err != nil {
			return nil, err
		}

		for _, replicationGroup := range resp.ReplicationGroups {
			// Arn creation for listing tags
		        tags, err := svc.ListTagsForResource(&elasticache.ListTagsForResourceInput{
			    ResourceName: aws.String(fmt.Sprintf("arn:aws-cn:elasticache:%s:%s:cluster:%s", *region, *accountID, *cacheCluster.CacheClusterId)),
		        })
		        if err != nil {
			    continue
		        }
			resources = append(resources, &ElasticacheReplicationGroup{
				svc:     svc,
				groupID: replicationGroup.ReplicationGroupId,
				tags:      tags.TagList,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

func (i *ElasticacheReplicationGroup) Remove() error {
	params := &elasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: i.groupID,
	}

	_, err := i.svc.DeleteReplicationGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheReplicationGroup) String() string {
	return *i.groupID
}
func (i *ElasticacheCacheCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Identifier", i.clusterID)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
