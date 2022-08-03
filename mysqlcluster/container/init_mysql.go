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
	corev1 "k8s.io/api/core/v1"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// initMysql used for init-mysql container.
type initMysql struct {
	*mysqlcluster.MysqlCluster

	// The name of the init-mysql container.
	name string
}

// getName get the container name.
func (c *initMysql) getName() string {
	return c.name
}

// getImage get the container image.
func (c *initMysql) getImage() string {
	img := utils.MysqlImageVersions[c.GetMySQLVersion()]
	return img
}

// getCommand get the container command.
func (c *initMysql) getCommand() []string {
	// Because initialize mysql contain error, so do it in commands.
	return []string{"sh", "-c", "/docker-entrypoint.sh mysqld;if test -f /docker-entrypoint-initdb.d/plugin.sh; then /docker-entrypoint-initdb.d/plugin.sh; fi "}
}

// getEnvVars get the container env.
func (c *initMysql) getEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
			Value: "yes",
		},
		{
			Name:  "MYSQL_ROOT_HOST",
			Value: c.Spec.MysqlOpts.RootHost,
		},
		{
			Name:  "MYSQL_INIT_ONLY",
			Value: "1",
		},
	}

	sctName := c.GetNameForResource(utils.Secret)
	envs = append(
		envs,
		getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false),
	)

	if c.Spec.MysqlOpts.InitTokuDB {
		envs = append(envs, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envs
}

// getLifecycle get the container lifecycle.
func (c *initMysql) getLifecycle() *corev1.Lifecycle {
	return nil
}

// getResources get the container resources.
func (c *initMysql) getResources() corev1.ResourceRequirements {
	return c.Spec.MysqlOpts.Resources
}

// getPorts get the container ports.
func (c *initMysql) getPorts() []corev1.ContainerPort {
	return nil
}

// getProbeSet get the set of livenessProbe, ReadinessProbe and StartupProbe.
func (c *initMysql) getProbeSet() *ProbeSet {
	return &ProbeSet{
		LivenessProbe:  nil,
		ReadinessProbe: nil,
		StartupProbe:   nil,
	}
}

// getVolumeMounts get the container volumeMounts.
func (c *initMysql) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.MysqlConfVolumeName,
			MountPath: utils.MysqlConfVolumeMountPath,
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		},
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
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
}
