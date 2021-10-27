/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sidecar

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/blang/semver"
	"github.com/go-ini/ini"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// Config of the sidecar.
type Config struct {
	// The hostname of the pod.
	HostName string
	// The namespace where the pod is in.
	NameSpace string
	// The name of the headless service.
	ServiceName string
	// The name of the statefulset.
	StatefulSetName string
	// Replicas is the number of pods.
	Replicas int32

	// The password of the root user.
	RootPassword string

	// Username of new user to create.
	User string
	// Password for the new user.
	Password string
	// Name for new database to create.
	Database string

	// The name of replication user.
	ReplicationUser string
	// The password of the replication user.
	ReplicationPassword string

	// The name of metrics user.
	MetricsUser string
	// The password of metrics user.
	MetricsPassword string

	// The name of operator user.
	OperatorUser string
	// The password of operator user.
	OperatorPassword string
	// The password of the mysql root user, for operator use only.
	InternalRootPassword string

	// InitTokuDB represents if install tokudb engine.
	InitTokuDB bool

	// MySQLVersion represents the MySQL version that will be run.
	MySQLVersion semver.Version

	// The parameter in xenon means admit defeat count for hearbeat.
	AdmitDefeatHearbeatCount int32
	// The parameter in xenon means election timeout(ms)。
	ElectionTimeout int32

	// Whether the MySQL data exists.
	existMySQLData bool
	// for mysql backup
	// backup user and password for http endpoint
	ClusterName string

	// Backup user name to http Server
	BackupUser string

	// Backup Password to htpp Server
	BackupPassword string

	// XbstreamExtraArgs is a list of extra command line arguments to pass to xbstream.
	XbstreamExtraArgs []string

	// XtrabackupExtraArgs is a list of extra command line arguments to pass to xtrabackup.
	XtrabackupExtraArgs []string

	// XtrabackupPrepareExtraArgs is a list of extra command line arguments to pass to xtrabackup
	// during --prepare.
	XtrabackupPrepareExtraArgs []string

	// XtrabackupTargetDir is a backup destination directory for xtrabackup.
	XtrabackupTargetDir string

	// S3 endpoint domain name
	XCloudS3EndPoint string

	// S3 access key
	XCloudS3AccessKey string

	// S3 secrete key
	XCloudS3SecretKey string

	// S3 Bucket names
	XCloudS3Bucket string

	// directory in S3 bucket for cluster restore from
	XRestoreFrom string
}

// NewInitConfig returns the configuration file needed for initialization.
func NewInitConfig() *Config {
	mysqlVersion, err := semver.Parse(getEnvValue("MYSQL_VERSION"))
	if err != nil {
		log.Info("MYSQL_VERSION is not a semver version")
		mysqlVersion, _ = semver.Parse(utils.MySQLDefaultVersion)
	}

	replicaStr := getEnvValue("REPLICAS")
	replicas, err := strconv.ParseInt(replicaStr, 10, 32)
	if err != nil {
		log.Error(err, "invalid environment values", "REPLICAS", replicaStr)
		panic(err)
	}

	initTokuDB := false
	if len(getEnvValue("INIT_TOKUDB")) > 0 {
		initTokuDB = true
	}

	admitDefeatHearbeatCount, err := strconv.ParseInt(getEnvValue("ADMIT_DEFEAT_HEARBEAT_COUNT"), 10, 32)
	if err != nil {
		admitDefeatHearbeatCount = 5
	}
	electionTimeout, err := strconv.ParseInt(getEnvValue("ELECTION_TIMEOUT"), 10, 32)
	if err != nil {
		electionTimeout = 10000
	}

	existMySQLData, _ := checkIfPathExists(fmt.Sprintf("%s/mysql", dataPath))

	return &Config{
		HostName:        getEnvValue("POD_HOSTNAME"),
		NameSpace:       getEnvValue("NAMESPACE"),
		ServiceName:     getEnvValue("SERVICE_NAME"),
		StatefulSetName: getEnvValue("STATEFULSET_NAME"),
		Replicas:        int32(replicas),

		RootPassword:         getEnvValue("MYSQL_ROOT_PASSWORD"),
		InternalRootPassword: getEnvValue("INTERNAL_ROOT_PASSWORD"),

		Database: getEnvValue("MYSQL_DATABASE"),
		User:     getEnvValue("MYSQL_USER"),
		Password: getEnvValue("MYSQL_PASSWORD"),

		ReplicationUser:     getEnvValue("MYSQL_REPL_USER"),
		ReplicationPassword: getEnvValue("MYSQL_REPL_PASSWORD"),

		MetricsUser:     getEnvValue("METRICS_USER"),
		MetricsPassword: getEnvValue("METRICS_PASSWORD"),

		OperatorUser:     getEnvValue("OPERATOR_USER"),
		OperatorPassword: getEnvValue("OPERATOR_PASSWORD"),

		InitTokuDB: initTokuDB,

		MySQLVersion: mysqlVersion,

		AdmitDefeatHearbeatCount: int32(admitDefeatHearbeatCount),
		ElectionTimeout:          int32(electionTimeout),

		existMySQLData:    existMySQLData,
		XRestoreFrom:      getEnvValue("RESTORE_FROM"),
		XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
	}
}

// NewBackupConfig returns the configuration file needed for backup container.
func NewBackupConfig() *Config {
	replicaStr := getEnvValue("REPLICAS")
	replicas, err := strconv.ParseInt(replicaStr, 10, 32)
	if err != nil {
		log.Error(err, "invalid environment values", "REPLICAS", replicaStr)
		panic(err)
	}

	return &Config{
		NameSpace:    getEnvValue("NAMESPACE"),
		ServiceName:  getEnvValue("SERVICE_NAME"),
		Replicas:     int32(replicas),
		ClusterName:  getEnvValue("SERVICE_NAME"),
		RootPassword: getEnvValue("MYSQL_ROOT_PASSWORD"),

		BackupUser:     getEnvValue("BACKUP_USER"),
		BackupPassword: getEnvValue("BACKUP_PASSWORD"),

		XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
	}
}

// NewReqBackupConfig returns the configuration file needed for backup job.
func NewReqBackupConfig() *Config {
	replicaStr := getEnvValue("REPLICAS")
	replicas, err := strconv.ParseInt(replicaStr, 10, 32)
	if err != nil {
		log.Error(err, "invalid environment values", "REPLICAS", replicaStr)
		panic(err)
	}

	return &Config{
		NameSpace:   getEnvValue("NAMESPACE"),
		ServiceName: getEnvValue("SERVICE_NAME"),
		Replicas:    int32(replicas),

		BackupUser:     getEnvValue("BACKUP_USER"),
		BackupPassword: getEnvValue("BACKUP_PASSWORD"),
	}
}

// GetContainerType returns the CONTAINER_TYPE of the currently running container.
// CONTAINER_TYPE used to mark the container type.
func GetContainerType() string {
	return getEnvValue("CONTAINER_TYPE")
}

// build Xtrabackup arguments
func (cfg *Config) XtrabackupArgs() []string {
	// xtrabackup --backup <args> --target-dir=<backup-dir> <extra-args>
	tmpdir := "/root/backup/"
	if len(cfg.XtrabackupTargetDir) != 0 {
		tmpdir = cfg.XtrabackupTargetDir
	}
	xtrabackupArgs := []string{
		"--backup",
		"--stream=xbstream",
		"--host=127.0.0.1",
		fmt.Sprintf("--user=%s", utils.RootUser),
		fmt.Sprintf("--password=%s", cfg.RootPassword),
		fmt.Sprintf("--target-dir=%s", tmpdir),
	}

	return append(xtrabackupArgs, cfg.XtrabackupExtraArgs...)
}

// Build xbcloud arguments
func (cfg *Config) XCloudArgs() []string {
	xcloudArgs := []string{
		"put",
		"--storage=S3",
		fmt.Sprintf("--s3-endpoint=%s", cfg.XCloudS3EndPoint),
		fmt.Sprintf("--s3-access-key=%s", cfg.XCloudS3AccessKey),
		fmt.Sprintf("--s3-secret-key=%s", cfg.XCloudS3SecretKey),
		fmt.Sprintf("--s3-bucket=%s", cfg.XCloudS3Bucket),
		"--parallel=10",
		utils.BuildBackupName(),
		"--insecure",
	}
	return xcloudArgs
}

// buildExtraConfig build a ini file for mysql.
func (cfg *Config) buildExtraConfig(filePath string) (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("mysqld")

	ordinal, err := utils.GetOrdinal(cfg.HostName)
	if err != nil {
		return nil, err
	}
	if _, err := sec.NewKey("server-id", strconv.Itoa(mysqlServerIDOffset+ordinal)); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("init-file", filePath); err != nil {
		return nil, err
	}

	return conf, nil
}

// buildXenonConf build a config file for xenon.
func (cfg *Config) buildXenonConf() []byte {
	pingTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	heartbeatTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	requestTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount

	version := "mysql80"
	if cfg.MySQLVersion.Major == 5 {
		if cfg.MySQLVersion.Minor == 6 {
			version = "mysql56"
		} else {
			version = "mysql57"
		}
	}

	var masterSysVars, slaveSysVars string
	if cfg.InitTokuDB {
		masterSysVars = "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	} else {
		masterSysVars = "sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	}

	hostName := fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)

	str := fmt.Sprintf(`{
    "log": {
        "level": "INFO"
    },
    "server": {
        "endpoint": "%s:%d",
        "peer-address": "%s:%d",
        "enable-apis": true
    },
    "replication": {
        "passwd": "%s",
        "user": "%s"
    },
    "rpc": {
        "request-timeout": %d
    },
    "mysql": {
        "admit-defeat-ping-count": 3,
        "admin": "root",
        "ping-timeout": %d,
        "passwd": "%s",
        "host": "localhost",
        "version": "%s",
        "master-sysvars": "%s",
        "slave-sysvars": "%s",
        "port": 3306,
        "monitor-disabled": true
    },
    "raft": {
        "election-timeout": %d,
        "admit-defeat-hearbeat-count": %d,
        "heartbeat-timeout": %d,
        "meta-datadir": "/var/lib/xenon/",
        "leader-start-command": "/scripts/leader-start.sh",
        "leader-stop-command": "/scripts/leader-stop.sh",
        "semi-sync-degrade": true,
        "purge-binlog-disabled": true,
        "super-idle": false
    }
}
`, hostName, utils.XenonPort, hostName, utils.XenonPeerPort, cfg.ReplicationPassword, cfg.ReplicationUser, requestTimeout,
		pingTimeout, cfg.RootPassword, version, masterSysVars, slaveSysVars, cfg.ElectionTimeout,
		cfg.AdmitDefeatHearbeatCount, heartbeatTimeout)
	return utils.StringToBytes(str)
}

// buildInitSql used to build init.sql. The file run after the mysql init.
func (cfg *Config) buildInitSql() []byte {
	sql := fmt.Sprintf(`SET @@SESSION.SQL_LOG_BIN=0;
CREATE DATABASE IF NOT EXISTS %s;
DROP user IF EXISTS 'root'@'127.0.0.1';
GRANT ALL ON *.* TO 'root'@'127.0.0.1' IDENTIFIED BY '%s' with grant option;
DROP user IF EXISTS 'root'@'%%';
GRANT ALL ON *.* TO 'root'@'%%' IDENTIFIED BY '%s' with grant option;
DROP user IF EXISTS '%s'@'%%';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO '%s'@'%%' IDENTIFIED BY '%s';
DROP user IF EXISTS '%s'@'%%';
GRANT SELECT, PROCESS, REPLICATION CLIENT ON *.* TO '%s'@'%%' IDENTIFIED BY '%s';
DROP user IF EXISTS '%s'@'%%';
GRANT SUPER, PROCESS, RELOAD, CREATE, SELECT ON *.* TO '%s'@'%%' IDENTIFIED BY '%s';
DROP user IF EXISTS '%s'@'%%';
GRANT ALL ON %s.* TO '%s'@'%%' IDENTIFIED BY '%s';
FLUSH PRIVILEGES;
`, cfg.Database, cfg.RootPassword, cfg.InternalRootPassword, cfg.ReplicationUser, cfg.ReplicationUser, cfg.ReplicationPassword,
		cfg.MetricsUser, cfg.MetricsUser, cfg.MetricsPassword, cfg.OperatorUser, cfg.OperatorUser,
		cfg.OperatorPassword, cfg.User, cfg.Database, cfg.User, cfg.Password)

	return utils.StringToBytes(sql)
}

// buildClientConfig used to build client.conf.
func (cfg *Config) buildClientConfig() (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("client")

	if _, err := sec.NewKey("host", "127.0.0.1"); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("port", fmt.Sprintf("%d", utils.MysqlPort)); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("user", cfg.OperatorUser); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("password", cfg.OperatorPassword); err != nil {
		return nil, err
	}

	return conf, nil
}

func (cfg *Config) buildPostStart() ([]byte, error) {
	ordinal, err := utils.GetOrdinal(cfg.HostName)
	if err != nil {
		return nil, err
	}

	nums := ordinal
	if cfg.existMySQLData {
		nums = int(cfg.Replicas)
	}

	host := fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)

	str := fmt.Sprintf(`#!/bin/sh
while true; do
	info=$(curl -i -X GET -u root:%s http://%s:%d/v1/xenon/ping)
	code=$(echo $info|grep "HTTP"|awk '{print $2}')
	if [ "$code" -eq "200" ]; then
		break
	fi
done
`, cfg.RootPassword, host, utils.XenonPeerPort)

	if !cfg.existMySQLData && ordinal == 0 {
		str = fmt.Sprintf(`%s
for i in $(seq 12); do
	curl -i -X POST -u root:%s http://%s:%d/v1/raft/trytoleader
	sleep 5
	curl -i -X GET -u root:%s http://%s:%d/v1/raft/status | grep LEADER
	if [ $? -eq 0 ] ; then
		echo "trytoleader success"
		break
	fi
	if [ $i -eq 12 ]; then
		echo "wait trytoleader failed"
	fi
done
`, str, cfg.RootPassword, host, utils.XenonPeerPort, cfg.RootPassword, host, utils.XenonPeerPort)
	} else {
		str = fmt.Sprintf(`%s
i=0
while [ $i -lt %d ]; do
	if [ $i -ne %d ]; then
		for k in $(seq 12); do
			res=$(curl -i -X POST -d '{"address": "%s-'$i'.%s.%s:%d"}' -u root:%s http://%s:%d/v1/cluster/add)
			code=$(echo $res|grep "HTTP"|awk '{print $2}')
			if [ "$code" -eq "200" ]; then
				break
			fi
		done

		for k in $(seq 12); do
			res=$(curl -i -X POST -d '{"address": "%s:%d"}' -u root:%s http://%s-$i.%s.%s:%d/v1/cluster/add)
			code=$(echo $res|grep "HTTP"|awk '{print $2}')
			if [ "$code" -eq "200" ]; then
				break
			fi
		done
	fi
	i=$((i+1))
done
`, str, nums, ordinal, cfg.StatefulSetName, cfg.ServiceName, cfg.NameSpace, utils.XenonPort,
			cfg.RootPassword, host, utils.XenonPeerPort, host, utils.XenonPort, cfg.RootPassword,
			cfg.StatefulSetName, cfg.ServiceName, cfg.NameSpace, utils.XenonPeerPort)
	}

	return utils.StringToBytes(str), nil
}

func (cfg *Config) buildPreStop() []byte {
	host := fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)

	str := fmt.Sprintf(`#!/bin/sh
while true; do
	info=$(curl -i -X GET -u root:%s http://%s:%d/v1/xenon/ping)
	code=$(echo $info|grep "HTTP"|awk '{print $2}')
	if [ "$code" -eq "200" ]; then
		break
	fi
done

curl -i -X PUT -u root:%s http://%s:%d/v1/raft/disable
for line in $(curl -X GET -u root:%s http://%s:%d/v1/raft/status | jq -r .nodes[] | cut -d : -f 1)
do
	if [ "$line" != "%s" ]; then
		for i in $(seq 12); do
			info=$(curl -i -X POST -d '{"address": "%s:%d"}' -u root:%s http://$line:%d/v1/cluster/remove)
			code=$(echo $info|grep "HTTP"|awk '{print $2}')
			if [ "$code" -eq "200" ]; then
				break
			fi
			if [ $i -eq 12 ]; then
				echo "remove node failed"
				break
			fi
			sleep 5
		done
	fi
done
`, cfg.RootPassword, host, utils.XenonPeerPort, cfg.RootPassword, host, utils.XenonPeerPort, cfg.RootPassword,
		host, utils.XenonPeerPort, host, host, utils.XenonPort, cfg.RootPassword, utils.XenonPeerPort)

	return utils.StringToBytes(str)
}

// buildLeaderStart build the leader-start.sh.
func (cfg *Config) buildLeaderStart() []byte {
	str := fmt.Sprintf(`#!/usr/bin/env bash
curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
--cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
-d '[{"op": "replace", "path": "/metadata/labels/role", "value": "leader"}]'
`, cfg.NameSpace)
	return utils.StringToBytes(str)
}

// buildLeaderStop build the leader-stop.sh.
func (cfg *Config) buildLeaderStop() []byte {
	str := fmt.Sprintf(`#!/usr/bin/env bash
curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
--cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
-d '[{"op": "replace", "path": "/metadata/labels/role", "value": "follower"}]'
`, cfg.NameSpace)
	return utils.StringToBytes(str)
}

// build S3 restore shell script
func (cfg *Config) buildS3Restore(path string) error {
	if len(cfg.XRestoreFrom) == 0 {
		return fmt.Errorf("do not have restore from")
	}
	if len(cfg.XCloudS3EndPoint) == 0 ||
		len(cfg.XCloudS3AccessKey) == 0 ||
		len(cfg.XCloudS3SecretKey) == 0 ||
		len(cfg.XCloudS3Bucket) == 0 {
		return fmt.Errorf("do not have S3 information")
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create restore.sh fail : %s", err)
	}
	defer func() {
		f.Close()
	}()

	restoresh := `#!/bin/sh
if [ ! -d {{.DataDir}} ] ; then
    echo "is not exist the var lib mysql"
    mkdir {{.DataDir}} 
    chown -R mysql.mysql {{.DataDir}} 
fi
mkdir /root/backup
xbcloud get --storage=S3 \
--s3-endpoint="{{.XCloudS3EndPoint}}" \
--s3-access-key="{{.XCloudS3AccessKey}}" \
--s3-secret-key="{{.XCloudS3SecretKey}}" \
--s3-bucket="{{.XCloudS3Bucket}}" \
--parallel=10 {{.XRestoreFrom}} \
--insecure |xbstream -xv -C /root/backup
# prepare redolog
xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --apply-log-only --target-dir=/root/backup
# prepare data
xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --target-dir=/root/backup
chown -R mysql.mysql /root/backup
xtrabackup --defaults-file={{.MyCnfMountPath}} --datadir={{.DataDir}} --copy-back --target-dir=/root/backup
chown -R mysql.mysql {{.DataDir}}
rm -rf /root/backup	
`
	template_restore := template.New("restore.sh")
	template_restore, err = template_restore.Parse(restoresh)
	if err != nil {
		return err
	}
	err2 := template_restore.Execute(f, struct {
		Config
		DataDir        string
		MyCnfMountPath string
	}{
		*cfg,
		utils.DataVolumeMountPath,
		utils.ConfVolumeMountPath + "/my.cnf",
	})
	if err2 != nil {
		return err2
	}
	return nil
}
