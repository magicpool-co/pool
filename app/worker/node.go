package worker

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/pkg/aws/ec2"
	"github.com/magicpool-co/pool/pkg/aws/ecs"
	"github.com/magicpool-co/pool/pkg/aws/route53"
	"github.com/magicpool-co/pool/pkg/aws/sqs"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

const (
	backupPath = "s3://magicpool-devops-prod/node-backups/"
	zoneName   = "privatemagicpool.co"
)

func getNodeClusterName(chain, env string, mainnet bool) string {
	chain = strings.ToLower(chain)
	if mainnet {
		return chain + "-full-nodes-" + env
	}

	return chain + "-testnet-full-nodes-" + env
}

func getNodeBackupPath(chain string, mainnet bool) string {
	chain = strings.ToLower(chain)
	if mainnet {
		return backupPath + chain
	}

	return backupPath + chain + "-testnet"
}

func getNodeBackupCommands(s3Path string) []string {
	cmd := fmt.Sprintf("aws s3 --region eu-west-1 sync /data/ %s --only-show-errors --delete", s3Path)
	return []string{cmd}
}

func getNodeContainer(awsClient *aws.Client, zoneID, cluster, url string) (string, string, error) {
	ip, err := route53.GetARecordIPByName(awsClient, zoneID, url)
	if err != nil {
		return "", "", err
	}

	instanceID, err := ec2.GetInstanceIDByIP(awsClient, ip)
	if err != nil {
		return "", "", err
	}

	containerID, err := ecs.GetContainerByInstanceID(awsClient, cluster, instanceID)
	if err != nil {
		return "", "", err
	}

	return instanceID, containerID, nil
}

type NodeStatusJob struct {
	locker  *redislock.Client
	logger  *log.Logger
	pooldb  *dbcl.Client
	nodes   []types.MiningNode
	metrics *metrics.Client
}

func (j *NodeStatusJob) Run() {
	defer j.logger.RecoverPanic()
	// need to run even if locked since PingHosts sets the status of the hostpool.
	// if it isn't locked, don't update the database though since that isn't critical
	var didObtainLock bool
	lock, err := retrieveLock("cron:nodestatus", time.Minute*5, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
			return
		}
	} else {
		defer lock.Release(context.Background())
		didObtainLock = true
	}

	for _, node := range j.nodes {
		hostIDs, heights, syncings, errs := node.PingHosts()
		for i := range hostIDs {
			if errs[i] != nil {
				j.logger.Error(errs[i])
				continue
			}

			var region string
			parts := strings.Split(hostIDs[i], ".")
			if len(parts) == 5 {
				region = parts[2]
			}

			if didObtainLock {
				poolNode := &pooldb.Node{
					URL:    hostIDs[i],
					Active: true,
					Height: types.Uint64Ptr(heights[i]),
					Synced: !syncings[i],
				}

				cols := []string{"active", "synced", "height"}
				err := pooldb.UpdateNode(j.pooldb.Writer(), poolNode, cols)
				if err != nil {
					j.logger.Error(err)
				}
			}

			if j.metrics != nil {
				j.metrics.SetGauge("node_height", float64(heights[i]), hostIDs[i], node.Chain(), region)
			}
		}
	}
}

type NodeInstanceChangeJob struct {
	env      string
	mainnet  bool
	locker   *redislock.Client
	logger   *log.Logger
	aws      *aws.Client
	telegram *telegram.Client
}

func (j *NodeInstanceChangeJob) Run() {
	defer j.logger.RecoverPanic()
	lock, err := retrieveLock("cron:nodeinstancechange", time.Minute*5, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	prefix := "mainnet"
	if !j.mainnet {
		prefix = "testnet"
	}

	zoneID, err := route53.GetZoneIDByName(j.aws, zoneName)
	if err != nil {
		j.logger.Error(err)
		return
	}

	for _, region := range []string{"eu-west-1", "eu-central-1", "us-east-1", "us-west-2"} {
		client, err := aws.NewSession(region, "")
		if err != nil {
			j.logger.Error(err)
			continue
		}

		queue := fmt.Sprintf("%s-full-node-asg-events-%s", prefix, region)
		msgs, err := sqs.PopFromQueue(client, queue)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		for _, msg := range msgs {
			asg, ok := msg.Attributes["AutoScalingGroupName"]
			if !ok {
				j.logger.Error(fmt.Errorf("no asg name for node instance event"))
				continue
			}

			chain, ok := msg.Attributes["chain"]
			if !ok {
				chain = strings.Split(asg, "-")[0]
			}

			chain = strings.ToLower(chain)
			if !j.mainnet {
				chain += "-testnet"
			}

			var instanceIP string
			if instanceID, ok := msg.Attributes["EC2InstanceId"]; ok {
				instanceIP, _ = ec2.GetInstanceIPByID(client, instanceID)
			}

			var needsRebalance bool
			switch msg.Attributes["LifecycleTransition"] {
			case "autoscaling:TEST_NOTIFICATION":
			case "autoscaling:EC2_INSTANCE_LAUNCHING":
				needsRebalance = true
				j.telegram.NotifyNodeInstanceLaunched(chain, region, instanceIP)
			case "autoscaling:EC2_INSTANCE_TERMINATING":
				needsRebalance = true
				j.telegram.NotifyNodeInstanceTerminated(chain, region, instanceIP)
			default:
				j.logger.Error(fmt.Errorf("unknown node instance event: %s", msg.Attributes["LifecycleTransition"]))
				continue
			}

			if needsRebalance && asg != "" && chain != "" {
				ips, err := ec2.GetGroupInstanceIPs(client, asg)
				if err != nil {
					j.logger.Error(err)
					continue
				}

				sort.Strings(ips)

				records := make(map[string]string, len(ips))
				for i, ip := range ips {
					key := fmt.Sprintf("node-%d.%s.%s.privatemagicpool.co", i, chain, region)
					records[key] = ip
				}

				err = route53.UpdateARecords(j.aws, zoneID, records)
				if err != nil {
					j.logger.Error(err)
				}
			}

			err := sqs.DeleteFromQueue(client, queue, msg.ID)
			if err != nil {
				j.logger.Error(err)
			}
		}
	}
}

type NodeCheckJob struct {
	env     string
	mainnet bool
	locker  *redislock.Client
	logger  *log.Logger
	aws     *aws.Client
	pooldb  *dbcl.Client
}

func (j *NodeCheckJob) Run() {
	defer j.logger.RecoverPanic()
	lock, err := retrieveLock("cron:nodecheck", time.Minute*5, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	nodes, err := pooldb.GetEnabledNodes(j.pooldb.Reader(), j.mainnet)
	if err != nil {
		j.logger.Error(err)
		return
	}

	var backupPeriod = time.Hour * 24 * 7
	const volumeThreshold = 80
	for _, node := range nodes {
		if node.Backup {
			needsBackup := node.BackupAt == nil
			if !needsBackup {
				switch node.ChainID {
				case "KAS":
					needsBackup = time.Since(types.TimeValue(node.BackupAt)) >= time.Hour*24*2
				default:
					needsBackup = time.Since(types.TimeValue(node.BackupAt)) >= backupPeriod
				}
			}

			if needsBackup {
				node.NeedsBackup = true
				node.PendingBackup = true
			}
		}

		cols := []string{"needs_backup", "pending_backup"}
		err = pooldb.UpdateNode(j.pooldb.Writer(), node, cols)
		if err != nil {
			j.logger.Error(err)
			continue
		}
	}
}

type NodeBackupJob struct {
	env     string
	mainnet bool
	locker  *redislock.Client
	logger  *log.Logger
	aws     *aws.Client
	pooldb  *dbcl.Client
}

func (j *NodeBackupJob) Run() {
	defer j.logger.RecoverPanic()
	lock, err := retrieveLock("cron:nodebackup", time.Hour*4, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	pendingNodes, err := pooldb.GetPendingBackupNodes(j.pooldb.Reader(), j.mainnet)
	if err != nil {
		j.logger.Error(err)
		return
	}

	zoneID, err := route53.GetZoneIDByName(j.aws, zoneName)
	if err != nil {
		j.logger.Error(err)
		return
	}

	for _, node := range pendingNodes {
		cluster := getNodeClusterName(node.ChainID, j.env, j.mainnet)
		s3Path := getNodeBackupPath(node.ChainID, j.mainnet)
		cmds := getNodeBackupCommands(s3Path)

		instanceID, containerID, err := getNodeContainer(j.aws, zoneID, cluster, node.URL)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		err = ecs.DrainClusterContainerInstance(j.aws, cluster, containerID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		time.Sleep(time.Second * 15)

		commandID, err := ec2.SendCommandToInstance(j.aws, instanceID, cmds)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		_, err = ec2.WaitForCommand(j.aws, instanceID, commandID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		err = ecs.ActivateClusterContainerInstance(j.aws, cluster, containerID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		node.NeedsBackup = false
		node.PendingBackup = false
		node.BackupAt = types.TimePtr(time.Now())

		cols := []string{"needs_backup", "pending_backup", "backup_at"}
		err = pooldb.UpdateNode(j.pooldb.Writer(), node, cols)
		if err != nil {
			j.logger.Error(err)
			continue
		}
	}
}
