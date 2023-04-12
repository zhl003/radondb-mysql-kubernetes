package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type raftStatus struct {
	Leader string   `json:"leader"`
	State  string   `json:"state"`
	Nodes  []string `json:"nodes"`
}

type MySQLNode struct {
	PodName   string
	Namespace string
	Role      string
}

type ClusterCredentials struct {
	SecretName           string
	InternalRootPassword string
	MysqlPassword        string
	ReplicationPassword  string
	MetricsPassword      string
	OperatorPassword     string
	MySQLRootPassword    string
	BackupPassword       string
}

type BackupCredentials struct {
	XCloudS3EndPoint  string
	XCloudS3AccessKey string
	XCloudS3SecretKey string
	XCloudS3Bucket    string
}

func GetClientSet() (*kubernetes.Clientset, error) {
	// Creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
	}
	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	return clientset, nil
}

func PatchRoleLabelTo(n MySQLNode) error {
	// Creates the clientset
	clientset, err := GetClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}
	patch := fmt.Sprintf(`{"metadata":{"labels":{"role":"%s"}}}`, n.Role)
	_, err = clientset.CoreV1().Pods(n.Namespace).Patch(context.TODO(), n.PodName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch pod role label: %v", err)
	}
	return nil
}

func XenonPingMyself() error {
	args := []string{"xenon", "ping"}
	cmd := exec.Command("xenoncli", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec xenoncli xenon ping: %v", err)
	}
	return nil
}

func GetRaftStatus() *raftStatus {
	args := []string{"raft", "status"}
	cmd := exec.Command("xenoncli", args...)
	res, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to exec xenoncli raft status: %v", err)
	}
	raftStatus := raftStatus{}
	if err := json.Unmarshal(res, &raftStatus); err != nil {
		log.Fatalf("failed to unmarshal raft status: %v", err)
	}
	return &raftStatus
}

func GetRole() string {
	return GetRaftStatus().State
}

// Get key, value from given secret
func GetClusterCredentials(namespace string) (*ClusterCredentials, error) {
	clusterSecretName := fmt.Sprintf("%s-secret", getEnvValue("CLUSTER_NAME"))
	keys := []string{"root-password", "internal-root-password", "mysql-password", "replication-password", "metrics-password", "operator-password"}
	credentials := &ClusterCredentials{
		SecretName: clusterSecretName,
	}
	clientSet, err := GetClientSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.Background(), clusterSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %v", clusterSecretName, err)
	}
	for _, key := range keys {
		value, ok := secret.Data[key]
		if !ok {
			return nil, fmt.Errorf("%s not found in secret %s: %v", key, clusterSecretName, err)
		}
		switch key {
		case "root-password":
			credentials.MySQLRootPassword = string(value)
		case "internal-root-password":
			credentials.InternalRootPassword = string(value)
		case "mysql-password":
			credentials.MysqlPassword = string(value)
		case "replication-password":
			credentials.ReplicationPassword = string(value)
		case "metrics-password":
			credentials.MetricsPassword = string(value)
		case "operator-password":
			credentials.OperatorPassword = string(value)
		case "backup-password":
			credentials.BackupPassword = string(value)
		}

	}
	return credentials, nil
}

// getEnvValue get environment variable by the key.
func getEnvValue(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "NOP"
	}

	return value
}

func GetBackupCredentials(namespace string) (*BackupCredentials, error) {
	// get MySQL cluster
	BackupSecretName := getEnvValue("BACKUP_SECRET_NAME")
	keys := []string{"s3-endpoint", "s3-access-key", "s3-secret-key", "s3-bucket"}
	credentials := &BackupCredentials{}

	clientSet, err := GetClientSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.Background(), BackupSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %v", BackupSecretName, err)
	}
	for _, key := range keys {
		value, ok := secret.Data[key]
		if !ok {
			return nil, fmt.Errorf("%s not found in secret %s: %v", key, BackupSecretName, err)
		}
		switch key {
		case "s3-endpoint":
			credentials.XCloudS3EndPoint = string(value)
		case "s3-access-key":
			credentials.XCloudS3AccessKey = string(value)
		case "s3-secret-key":
			credentials.XCloudS3SecretKey = string(value)
		case "s3-bucket":
			credentials.XCloudS3Bucket = string(value)
		}
	}
	return credentials, nil

}
