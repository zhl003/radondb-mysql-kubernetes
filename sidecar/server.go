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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	// backupStatus http trailer
	backupStatusTrailer = "X-Backup-Status"

	// success string
	backupSuccessful = "Success"

	// failure string
	backupFailed = "Failed"

	serverPort           = utils.XBackupPort
	serverProbeEndpoint  = "/health"
	serverBackupEndpoint = "/xbackup"
	serverConnectTimeout = 5 * time.Second

	// DownLoad server url.
	serverBackupDownLoadEndpoint = "/download"
)

type server struct {
	cfg *Config
	http.Server
}

// Create new Http Server.
func newServer(cfg *Config, stop <-chan struct{}) *server {
	mux := http.NewServeMux()
	srv := &server{
		cfg: cfg,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", serverPort),
			Handler: mux,
		},
	}

	// Add handle functions.
	mux.HandleFunc(serverProbeEndpoint, srv.healthHandler)
	mux.Handle(serverBackupEndpoint, maxClients(http.HandlerFunc(srv.backupHandler), 1))

	mux.Handle(serverBackupDownLoadEndpoint,
		maxClients(http.HandlerFunc(srv.backupDownLoadHandler), 1))

	// Shutdown gracefully the http server.
	go func() {
		<-stop // wait for stop signal
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error(err, "failed to stop http server")
		}
	}()

	return srv
}

// nolint: unparam
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Error(err, "failed writing request")
	}
}

func (s *server) backupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("content-type", "text/json")
	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}
	backName, Datetime, err := RunTakeBackupCommand(s.cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		msg, _ := json.Marshal(utils.JsonResult{Status: backupSuccessful, BackupName: backName, Date: Datetime})
		w.Write(msg)
	}
}

// DownLoad handler.
func (s *server) backupDownLoadHandler(w http.ResponseWriter, r *http.Request) {

	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "HTTP server does not support streaming!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Trailer", backupStatusTrailer)

	// nolint: gosec
	xtrabackup := exec.Command(xtrabackupCommand, s.cfg.XtrabackupArgs()...)
	xtrabackup.Stderr = os.Stderr

	stdout, err := xtrabackup.StdoutPipe()
	if err != nil {
		log.Error(err, "failed to create stdout pipe")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	defer func() {
		// don't care
		_ = stdout.Close()
	}()

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, stdout); err != nil {
		log.Error(err, "failed to copy buffer")
		http.Error(w, "buffer copy failed", http.StatusInternalServerError)
		return
	}

	if err := xtrabackup.Wait(); err != nil {
		log.Error(err, "failed waiting for xtrabackup to finish")
		w.Header().Set(backupStatusTrailer, backupFailed)
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	// success
	w.Header().Set(backupStatusTrailer, backupSuccessful)
	flusher.Flush()
}

func (s *server) isAuthenticated(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	return ok && user == s.cfg.BackupUser && pass == s.cfg.BackupPassword
}

// maxClients limit an http endpoint to allow just n max concurrent connections.
func maxClients(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() {
			<-sema
		}()
		h.ServeHTTP(w, r)
	})
}

func prepareURL(svc string, endpoint string) string {
	if !strings.Contains(svc, ":") {
		svc = fmt.Sprintf("%s:%d", svc, serverPort)
	}
	return fmt.Sprintf("http://%s%s", svc, endpoint)
}

// Set the timeout for HTTP.
func transportWithTimeout(connectTimeout time.Duration) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func setAnnonations(cfg *Config, backname string, DateTime string, BackupType string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	job, err := clientset.BatchV1().Jobs(cfg.NameSpace).Get(context.TODO(), cfg.JobName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if job.Annotations == nil {
		job.Annotations = make(map[string]string)
	}
	job.Annotations[utils.JobAnonationName] = backname
	job.Annotations[utils.JobAnonationDate] = DateTime
	job.Annotations[utils.JobAnonationType] = BackupType
	_, err = clientset.BatchV1().Jobs(cfg.NameSpace).Update(context.TODO(), job, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// requestABackup connects to specified host and endpoint and gets the backup.
func requestABackup(cfg *Config, host string, endpoint string) (*http.Response, error) {
	log.Info("initialize a backup", "host", host, "endpoint", endpoint)

	req, err := http.NewRequest("GET", prepareURL(host, endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %s", err)
	}

	// set authentication user and password
	req.SetBasicAuth(cfg.BackupUser, cfg.BackupPassword)

	client := &http.Client{}
	client.Transport = transportWithTimeout(serverConnectTimeout)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		status := "unknown"
		if resp != nil {
			status = resp.Status
		}
		return nil, fmt.Errorf("fail to get backup: %s, code: %s", err, status)
	}
	defer resp.Body.Close()
	var result utils.JsonResult
	json.NewDecoder(resp.Body).Decode(&result)

	err = setAnnonations(cfg, result.BackupName, result.Date, "S3") // set annotation
	if err != nil {
		return nil, fmt.Errorf("fail to set annotation: %s", err)
	}
	return resp, nil
}
