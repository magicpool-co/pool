package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/aws"
	"github.com/magicpool-co/pool/pkg/aws/ec2"
	"github.com/magicpool-co/pool/pkg/aws/ecs"
	"github.com/magicpool-co/pool/pkg/aws/route53"
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

	var didObtainLock bool
	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:nodestatus", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
			return
		}
	} else {
		defer lock.Release(ctx)
		didObtainLock = true
	}

	for _, node := range j.nodes {
		hostIDs, heights, syncings, errs := node.PingHosts()
		for i := range hostIDs {
			if errs[i] != nil {
				j.logger.Error(fmt.Errorf("status: %v", err))
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

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:nodecheck", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	nodes, err := pooldb.GetEnabledNodes(j.pooldb.Reader(), j.mainnet)
	if err != nil {
		j.logger.Error(err)
		return
	}

	/*zoneID, err := route53.GetZoneIDByName(j.aws, zoneName)
	if err != nil {
		j.logger.Error(err)
		return
	}*/

	var backupPeriod = time.Hour * 24 * 7
	const volumeThreshold = 80
	for _, node := range nodes {
		/*cluster := getNodeClusterName(node.ChainID, j.env, j.mainnet)
		_, _, err := getNodeContainer(j.aws, zoneID, cluster, node.URL)
		if err != nil {
			j.logger.Error(err)
			continue
		}*/

		// check for backup
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

		// check for update
		/* tasks, err := ecs.GetTasksByContainer(j.aws, cluster, containerID)
		if err != nil {
			j.logger.Error(err)
			continue
		} else if len(tasks) > 0 {
			// @TODO: maybe just check for active deployments instead
			latestRevision, err := ecs.GetServiceRevision(j.aws, cluster, cluster+"-service")
			if err != nil {
				j.logger.Error(err)
				continue
			}

			activeRevision, err := ecs.GetTaskRevision(j.aws, cluster, tasks[0])
			if err != nil {
				j.logger.Error(err)
				continue
			}

			if activeRevision != latestRevision {
				node.NeedsUpdate = true
			}
		} */

		/*// check for resize
		cmds := []string{"df /dev/nvme0n1p1 | awk 'END{print $5;}'"}
		commandID, err := ec2.SendCommandToInstance(j.aws, instanceID, cmds)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		rawVolumeUsage, err := ec2.WaitForCommand(j.aws, instanceID, commandID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		rawVolumeUsage = strings.ReplaceAll(rawVolumeUsage, "%", "")
		rawVolumeUsage = strings.ReplaceAll(rawVolumeUsage, "\n", "")
		volumeUsage, err := strconv.ParseInt(rawVolumeUsage, 10, 64)
		if err != nil {
			j.logger.Error(err)
			continue
		} else if volumeUsage >= volumeThreshold {
			node.NeedsResize = true
		}*/

		cols := []string{"needs_backup", "pending_backup", "needs_update",
			"pending_update", "needs_resize", "pending_resize"}
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

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:nodebackup", time.Hour*4, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

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
		start := time.Now()

		j.logger.Info(fmt.Sprintf("backing up %s (%s)", node.URL, cluster))

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

		j.logger.Info(fmt.Sprintf("running command with id %s on instance %s", commandID, instanceID))
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
		err = pooldb.UpdateNode(j.pooldb.Writer(), node, []string{"needs_backup", "pending_backup", "backup_at"})
		if err != nil {
			j.logger.Error(err)
			continue
		}

		j.logger.Info(fmt.Sprintf("finished backing up %s (%s) in %s", node.URL, cluster, time.Since(start)))
	}
}

type NodeUpdateJob struct {
	env     string
	mainnet bool
	locker  *redislock.Client
	logger  *log.Logger
	aws     *aws.Client
	pooldb  *dbcl.Client
}

func (j *NodeUpdateJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:nodeupdate", time.Hour*4, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	pendingNodes, err := pooldb.GetPendingUpdateNodes(j.pooldb.Reader(), j.mainnet)
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
		start := time.Now()

		j.logger.Info(fmt.Sprintf("updating %s (%s)", node.URL, cluster))

		_, containerID, err := getNodeContainer(j.aws, zoneID, cluster, node.URL)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		err = ecs.DrainClusterContainerInstance(j.aws, cluster, containerID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		time.Sleep(time.Second * 10)

		err = ecs.ActivateClusterContainerInstance(j.aws, cluster, containerID)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		node.NeedsUpdate = false
		node.PendingUpdate = false
		err = pooldb.UpdateNode(j.pooldb.Writer(), node, []string{"needs_update", "pending_update"})
		if err != nil {
			j.logger.Error(err)
			continue
		}

		j.logger.Info(fmt.Sprintf("finished updating %s (%s) in %s", node.URL, cluster, time.Since(start)))
	}
}

type NodeResizeJob struct {
	env     string
	mainnet bool
	locker  *redislock.Client
	logger  *log.Logger
	aws     *aws.Client
	pooldb  *dbcl.Client
}

func (j *NodeResizeJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:noderesize", time.Hour*4, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	pendingNodes, err := pooldb.GetPendingResizeNodes(j.pooldb.Reader(), j.mainnet)
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
		start := time.Now()

		j.logger.Info(fmt.Sprintf("resizing %s (%s)", node.URL, cluster))

		instanceID, _, err := getNodeContainer(j.aws, zoneID, cluster, node.URL)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		volumeIDs, err := ec2.GetInstanceVolumes(j.aws, instanceID)
		if err != nil {
			j.logger.Error(err)
			continue
		} else if len(volumeIDs) != 1 {
			j.logger.Error(fmt.Errorf("unable to find volume for instance %s", instanceID))
			continue
		}

		currentSize, err := ec2.GetEBSVolumeSize(j.aws, volumeIDs[0])
		if err != nil {
			j.logger.Error(err)
			continue
		}

		// add 20% of space to the volume
		newSize := currentSize + currentSize/5

		err = ec2.ResizeInstanceVolume(j.aws, instanceID, newSize)
		if err != nil {
			j.logger.Error(err)
			continue
		}

		j.logger.Info(fmt.Sprintf("finished resizing %s (%s) in %s", node.URL, cluster, time.Since(start)))
	}
}
