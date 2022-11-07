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

package syncer

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/presslabs/controller-util/pkg/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	// The length of the secret string.
	rStrLen = 24
	// policyASCII is the list of acceptable characters from which to generate an ASCII password.
	policyASCII = `` +
		`()*+,-./` + `:;<=>?@` + `[]^_` + `{|}` +
		`ABCDEFGHIJKLMNOPQRSTUVWXYZ` +
		`abcdefghijklmnopqrstuvwxyz` +
		`0123456789`
)

// NewSecretSyncer returns secret syncer.
func NewSecretSyncer(cli client.Client, c *mysqlcluster.MysqlCluster) syncer.Interface {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.Secret),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", c.Unwrap(), secret, cli, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		secret.Data["operator-user"] = []byte(utils.OperatorUser)
		if err := addRandomPassword(secret.Data, "operator-password"); err != nil {
			return err
		}

		// xtrabackup http server user and password
		secret.Data["backup-user"] = []byte(utils.BackupUser)
		if err := addRandomPassword(secret.Data, "backup-password"); err != nil {
			return err
		}

		secret.Data["metrics-user"] = []byte(utils.MetricsUser)
		if err := addRandomPassword(secret.Data, "metrics-password"); err != nil {
			return err
		}

		if c.Spec.MetricsOpts.Enabled {
			dataSource := fmt.Sprintf("%s:%s@(127.0.0.1:3306)/", utils.MetricsUser, utils.BytesToString(secret.Data["metrics-password"]))
			secret.Data["data-source"] = []byte(dataSource)
		}

		secret.Data["replication-user"] = []byte(utils.ReplicationUser)
		if err := addRandomPassword(secret.Data, "replication-password"); err != nil {
			return err
		}

		if err := addRandomPassword(secret.Data, "internal-root-password"); err != nil {
			return err
		}

		// if c.Spec.MysqlOpts.RootHost != "localhost" && c.Spec.MysqlOpts.RootPassword == "" {
		// 	if err := addRandomPassword(secret.Data, "root-password"); err != nil {
		// 		return err
		// 	}
		// } else {
		// 	secret.Data["root-password"] = []byte(c.Spec.MysqlOpts.RootPassword)
		// }
		// for security, specify root password in cr is not allowed
		if err := addRandomPassword(secret.Data, "root-password"); err != nil {
			return err
		}

		secret.Data["mysql-user"] = []byte(c.Spec.MysqlOpts.User)
		secret.Data["mysql-password"] = []byte(c.Spec.MysqlOpts.Password)
		secret.Data["mysql-database"] = []byte(c.Spec.MysqlOpts.Database)
		return nil
	})
}

// addRandomPassword checks if a key exists and if not registers a random string for that key
func addRandomPassword(data map[string][]byte, key string) error {
	if len(data[key]) == 0 {
		// NOTE: use only alpha-numeric string, this strings are used unescaped in MySQL queries.
		random, err := GenerateASCIIPassword(rStrLen)
		if err != nil {
			return err
		}
		data[key] = []byte(random)
	}
	return nil
}

// GenerateASCIIPassword returns a random string of printable ASCII characters.
func GenerateASCIIPassword(length int) (string, error) {
	var randomASCII = randomCharacter(rand.Reader, policyASCII)
	return accumulate(length, randomASCII)
}

// accumulate gathers n bytes from f and returns them as a string. It returns
// an empty string when f returns an error.
func accumulate(n int, f func() (byte, error)) (string, error) {
	result := make([]byte, n)

	for i := range result {
		if b, err := f(); err == nil {
			result[i] = b
		} else {
			return "", err
		}
	}

	return string(result), nil
}

// randomCharacter builds a function that returns random bytes from class.
func randomCharacter(random io.Reader, class string) func() (byte, error) {
	if random == nil {
		panic("requires a random source")
	}
	if len(class) == 0 {
		panic("class cannot be empty")
	}

	size := big.NewInt(int64(len(class)))

	return func() (byte, error) {
		if i, err := rand.Int(random, size); err == nil {
			return class[int(i.Int64())], nil
		} else {
			return 0, err
		}
	}
}

func NewMyClientConfSyncer(cli client.Client, c *mysqlcluster.MysqlCluster) syncer.Interface {
	// first get the main secret
	clusterSecret := &corev1.Secret{}
	if err := cli.Get(context.TODO(), types.NamespacedName{
		Name:      c.GetNameForResource(utils.Secret),
		Namespace: c.Namespace,
	}, clusterSecret); err != nil {
		return nil
	}
	password := utils.NewMyclientCnfFromSecret(clusterSecret)
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.ClientSecret),
			Namespace: c.Namespace,
		},
	}
	return syncer.NewObjectSyncer("Secret", c.Unwrap(), secret, cli, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data = password.ToMyClientCnfSecret().Data
		return nil
	})
}
