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

package container

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	defeatCount             int32 = 1
	electionTimeout         int32 = 5
	replicas                int32 = 3
	initSidecarMysqlCluster       = mysqlv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample",
		},
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			Replicas: &replicas,
			PodPolicy: mysqlv1alpha1.PodPolicy{
				SidecarImage: "sidecar image",
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			XenonOpts: mysqlv1alpha1.XenonOpts{
				AdmitDefeatHearbeatCount: &defeatCount,
				ElectionTimeout:          &electionTimeout,
			},
			MetricsOpts: mysqlv1alpha1.MetricsOpts{
				Enabled: false,
			},
			MysqlVersion: "5.7",
			MysqlOpts: mysqlv1alpha1.MysqlOpts{
				InitTokuDB: false,
			},
			Persistence: mysqlv1alpha1.Persistence{
				Enabled: false,
			},
		},
	}
	testInitSidecarCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &initSidecarMysqlCluster,
	}
	defaultInitSidecarEnvs = []corev1.EnvVar{
		{
			Name:  "CONTAINER_TYPE",
			Value: utils.ContainerInitSidecarName,
		},
		{
			Name: "POD_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  "NAMESPACE",
			Value: testInitSidecarCluster.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: "sample-mysql",
		},
		{
			Name:  "STATEFULSET_NAME",
			Value: "sample-mysql",
		},
		{
			Name:  "ADMIT_DEFEAT_HEARBEAT_COUNT",
			Value: strconv.Itoa(int(*testInitSidecarCluster.Spec.XenonOpts.AdmitDefeatHearbeatCount)),
		},
		{
			Name:  "ELECTION_TIMEOUT",
			Value: strconv.Itoa(int(*testInitSidecarCluster.Spec.XenonOpts.ElectionTimeout)),
		},
		{
			Name:  "MYSQL_VERSION",
			Value: "5.7.34",
		},
		{
			Name:  "RESTORE_FROM",
			Value: "",
		},
		{
			Name:  "CLUSTER_NAME",
			Value: "sample",
		},
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "root-password",
					Optional: &optFalse,
				},
			},
		},
		{
			Name: "INTERNAL_ROOT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "internal-root-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-database",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_REPL_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "replication-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_REPL_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "replication-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "METRICS_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "metrics-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "METRICS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "metrics-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "OPERATOR_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "operator-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "OPERATOR_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "operator-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "BACKUP_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "backup-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "BACKUP_PASSWORD",

			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "backup-password",
					Optional: &optTrue,
				},
			},
		},
	}
	defaultInitsidecarVolumeMounts = []corev1.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.ConfMapVolumeName,
			MountPath: utils.ConfMapVolumeMountPath,
		},
		{
			Name:      utils.ScriptsVolumeName,
			MountPath: utils.ScriptsVolumeMountPath,
		},
		{
			Name:      utils.XenonVolumeName,
			MountPath: utils.XenonVolumeMountPath,
		},
		{
			Name:      utils.InitFileVolumeName,
			MountPath: utils.InitFileVolumeMountPath,
		},
		{
			Name:      utils.SysLocalTimeZone,
			MountPath: utils.SysLocalTimeZoneMountPath,
		},
	}
	initSidecarCase = EnsureContainer("init-sidecar", &testInitSidecarCluster)
)

func TestGetInitSidecarName(t *testing.T) {
	assert.Equal(t, "init-sidecar", initSidecarCase.Name)
}

func TestGetInitSidecarImage(t *testing.T) {
	assert.Equal(t, "sidecar image", initSidecarCase.Image)
}

func TestGetInitSidecarCommand(t *testing.T) {
	command := []string{"sidecar", "init"}
	assert.Equal(t, command, initSidecarCase.Command)
}

func TestGetInitSidecarEnvVar(t *testing.T) {
	// default
	{
		assert.Equal(t, defaultInitSidecarEnvs, initSidecarCase.Env)
	}
	// initTokuDB
	{
		testToKuDBMysqlCluster := initSidecarMysqlCluster
		testToKuDBMysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testTokuDBCluster := mysqlcluster.MysqlCluster{
			MysqlCluster: &testToKuDBMysqlCluster,
		}
		tokudbCase := EnsureContainer("init-sidecar", &testTokuDBCluster)
		testTokuDBEnv := make([]corev1.EnvVar, len(defaultInitSidecarEnvs))
		copy(testTokuDBEnv, defaultInitSidecarEnvs)
		testTokuDBEnv = append(testTokuDBEnv, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
		assert.Equal(t, testTokuDBEnv, tokudbCase.Env)
	}
	// BackupSecretName not empty
	{
		testBackupMysqlCluster := initSidecarMysqlCluster
		testBackupMysqlCluster.Spec.BackupSecretName = "backup-secret"
		testBackupMysqlClusterWraper := mysqlcluster.MysqlCluster{
			MysqlCluster: &testBackupMysqlCluster,
		}
		BackupCase := EnsureContainer("init-sidecar", &testBackupMysqlClusterWraper)
		testBackupEnv := make([]corev1.EnvVar, len(defaultInitSidecarEnvs))
		copy(testBackupEnv, defaultInitSidecarEnvs)
		testBackupEnv = append(testBackupEnv,
			corev1.EnvVar{
				Name: "S3_ENDPOINT",

				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testBackupMysqlClusterWraper.Spec.BackupSecretName,
						},
						Key:      "s3-endpoint",
						Optional: &optFalse,
					},
				},
			},
			corev1.EnvVar{
				Name: "S3_ACCESSKEY",

				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testBackupMysqlClusterWraper.Spec.BackupSecretName,
						},
						Key:      "s3-access-key",
						Optional: &optTrue,
					},
				},
			},
			corev1.EnvVar{
				Name: "S3_SECRETKEY",

				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testBackupMysqlClusterWraper.Spec.BackupSecretName,
						},
						Key:      "s3-secret-key",
						Optional: &optTrue,
					},
				},
			},
			corev1.EnvVar{
				Name: "S3_BUCKET",

				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testBackupMysqlClusterWraper.Spec.BackupSecretName,
						},
						Key:      "s3-bucket",
						Optional: &optTrue,
					},
				},
			},
		)
		assert.Equal(t, testBackupEnv, BackupCase.Env)
	}
}

func TestGetInitSidecarLifecycle(t *testing.T) {
	assert.Nil(t, initSidecarCase.Lifecycle)
}

func TestGetInitSidecarResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, initSidecarCase.Resources)
}

func TestGetInitSidecarPorts(t *testing.T) {
	assert.Nil(t, initSidecarCase.Ports)
}

func TestGetInitSidecarLivenessProbe(t *testing.T) {
	assert.Nil(t, initSidecarCase.LivenessProbe)
}

func TestGetInitSidecarReadinessProbe(t *testing.T) {
	assert.Nil(t, initSidecarCase.ReadinessProbe)
}

func TestGetInitSidecarVolumeMounts(t *testing.T) {
	// default
	{
		assert.Equal(t, defaultInitsidecarVolumeMounts, initSidecarCase.VolumeMounts)
	}
	// init tokudb
	{
		testToKuDBMysqlCluster := initSidecarMysqlCluster
		testToKuDBMysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testTokuDBCluster := mysqlcluster.MysqlCluster{
			MysqlCluster: &testToKuDBMysqlCluster,
		}
		tokudbCase := EnsureContainer("init-sidecar", &testTokuDBCluster)
		tokuDBVolumeMounts := make([]corev1.VolumeMount, 6, 7)
		copy(tokuDBVolumeMounts, defaultInitsidecarVolumeMounts)
		tokuDBVolumeMounts = append(tokuDBVolumeMounts, corev1.VolumeMount{
			Name:      utils.SysVolumeName,
			MountPath: utils.SysVolumeMountPath,
		})
		assert.Equal(t, tokuDBVolumeMounts, tokudbCase.VolumeMounts)
	}
	// enable persistence
	{
		testPersistenceMysqlCluster := initSidecarMysqlCluster
		testPersistenceMysqlCluster.Spec.Persistence.Enabled = true
		testPersistenceCluster := mysqlcluster.MysqlCluster{
			MysqlCluster: &testPersistenceMysqlCluster,
		}
		persistenceCase := EnsureContainer("init-sidecar", &testPersistenceCluster)
		persistenceVolumeMounts := make([]corev1.VolumeMount, 6, 7)
		copy(persistenceVolumeMounts, defaultInitsidecarVolumeMounts)
		persistenceVolumeMounts = append(persistenceVolumeMounts, corev1.VolumeMount{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		})
		assert.Equal(t, persistenceVolumeMounts, persistenceCase.VolumeMounts)
	}
}
