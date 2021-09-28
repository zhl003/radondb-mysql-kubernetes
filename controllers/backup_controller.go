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

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/presslabs/controller-util/syncer"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/backup"
	backupSyncer "github.com/radondb/radondb-mysql-kubernetes/backup/syncer"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// BackupReconciler reconciles a Backup object.
type BackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=mysql.radondb.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=backups/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// your logic here
	// Fetch the Backup instance
	backup := backup.New(&apiv1alpha1.Backup{})
	err := r.Get(context.TODO(), req.NamespacedName, backup.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// Set defaults on backup
	r.Scheme.Default(backup.Unwrap())

	// save the backup for later check for diff
	savedBackup := backup.Unwrap().DeepCopy()

	jobSyncer := backupSyncer.NewJobSyncer(r.Client, r.Scheme, backup)
	if err := syncer.Sync(ctx, jobSyncer, r.Recorder); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.updateBackup(savedBackup, backup); err != nil {
		return reconcile.Result{}, err
	}

	// Clear the backup, Just keep historyLimit len
	if err = r.clearHistoryJob(ctx, req, *backup.Spec.HistoryLimit); err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

// Clear the History finished Jobs over HistoryLimit.
func (r *BackupReconciler) clearHistoryJob(ctx context.Context, req ctrl.Request, historyLimit int32) error {
	log := log.Log.WithName("controllers").WithName("Backup")
	backupJobs := batchv1.JobList{}
	labelSet := labels.Set{"Type": utils.BackupJobTypeName}
	if err := r.List(context.TODO(), &backupJobs, &client.ListOptions{
		Namespace: req.Namespace, LabelSelector: labelSet.AsSelector(),
	}); err != nil {
		return err
	}

	var finishedBackups []*batchv1.Job
	for _, job := range backupJobs.Items {
		if IsJobFinished(&job) {
			finishedBackups = append(finishedBackups, &job)
		}
	}

	sort.Slice(finishedBackups, func(i, j int) bool {
		if finishedBackups[i].Status.StartTime == nil {
			return finishedBackups[j].Status.StartTime != nil
		}
		return finishedBackups[i].Status.StartTime.Before(finishedBackups[j].Status.StartTime)
	})

	for i, job := range finishedBackups {
		if int32(i) >= int32(len(finishedBackups))-historyLimit {
			break
		}
		// at first check backup status completed.
		backup := backup.New(&apiv1alpha1.Backup{})
		namespacedName := types.NamespacedName{
			Name:      strings.TrimSuffix(job.Name, "-backup"),
			Namespace: job.Namespace,
		}
		if err := r.Get(context.TODO(), namespacedName, backup.Unwrap()); err != nil {
			log.Error(err, "can not find the backup", "jobName", job.Name)
			break
		}
		if backup.Status.Completed {
			if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete old completed job", "job", job)
			} else {
				log.V(0).Info("deleted old completed job", "job", job)
			}
		}

	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Backup{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

// Update backup Object and Status.
func (r *BackupReconciler) updateBackup(savedBackup *apiv1alpha1.Backup, backup *backup.Backup) error {
	log := log.Log.WithName("controllers").WithName("Backup")
	if !reflect.DeepEqual(savedBackup, backup.Unwrap()) {
		if err := r.Update(context.TODO(), backup.Unwrap()); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(savedBackup.Status, backup.Unwrap().Status) {
		log.Info("update backup object status")
		if err := r.Status().Update(context.TODO(), backup.Unwrap()); err != nil {
			log.Error(err, fmt.Sprintf("update status backup %s/%s", backup.Name, backup.Namespace),
				"backupStatus", backup.Status, "saveBackupStatus", savedBackup.Status)
			return err
		}
	}
	return nil
}

// IsJobFinished checks whether the given Job has finished execution.
// It does not discriminate between successful and failed terminations.
// See https://github.com/kubernetes/kubernetes/blob/7a0638da76cb9843def65708b661d2c6aa58ed5a/pkg/controller/job/utils.go#L26
func IsJobFinished(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
