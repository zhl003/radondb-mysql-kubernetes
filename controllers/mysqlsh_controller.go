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
	"io"
	"reflect"

	"github.com/pkg/errors"
	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	"go.opentelemetry.io/otel/trace"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ConditionRepoHostReady = "MySQLshRepoHostReady"
)

// MySQLshReconciler reconciles a Backup object.
type MySQLshReconciler struct {
	Client   client.Client
	Owner    client.FieldOwner
	Recorder record.EventRecorder
	Tracer   trace.Tracer
	PodExec  func(
		namespace, pod, container string,
		stdin io.Reader, stdout, stderr io.Writer, command ...string,
	) error
}

// +kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MySQLshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the cluster instance
	// create the result that will be updated following a call to each reconciler
	// ctx, span := r.Tracer.Start(ctx, "Reconcile")
	log := log.Log.WithName("controllers").WithName("MySQLsh")
	result := reconcile.Result{}

	updateResult := func(next reconcile.Result, err error) error {
		if err == nil {
			result = updateReconcileResult(result, next)
		}
		return err
	}
	cluster := mysqlcluster.New(&apiv1alpha1.MysqlCluster{})
	// get the mysqlcluster from the cache
	if err := r.Client.Get(ctx, req.NamespacedName, cluster.Unwrap()); err != nil {
		// NotFound cannot be fixed by requeuing so ignore it. During background
		// deletion, we receive delete events from cluster's dependents after
		// cluster is deleted.
		if err = client.IgnoreNotFound(err); err != nil {
			log.Error(err, "unable to fetch MySQLCluster")
			// span.RecordError(err)
		}
		return result, err
	}
	before := cluster.DeepCopy()
	var err error
	// Define the function for the updating the MySQLCluster status. Returns any error that
	// occurs while attempting to patch the status, while otherwise simply returning the
	// Result and error variables that are populated while reconciling the MySQLCluster.
	patchClusterStatus := func() (reconcile.Result, error) {
		if !equality.Semantic.DeepEqual(before.Status, cluster.Status) {
			// NOTE(cbandy): Kubernetes prior to v1.16.10 and v1.17.6 does not track
			// managed fields on the status subresource: https://issue.k8s.io/88901
			if err := errors.WithStack(r.Client.Status().Patch(
				ctx, cluster.MysqlCluster, client.MergeFrom(before), r.Owner)); err != nil {
				log.Error(err, "patching cluster status")
				return result, err
			}
			log.V(1).Info("patched cluster status")
		}
		return result, err
	}
	if err == nil {
		err = updateResult(r.reconcileMySQLsh(ctx, cluster))
	}

	return patchClusterStatus()
}

func (r *MySQLshReconciler) reconcileMySQLsh(ctx context.Context,
	cluster *mysqlcluster.MysqlCluster) (reconcile.Result, error) {
	// log := log.Log.WithName("controllers").WithName("MySQLsh")

	// if nil create the mysqlsh status
	// if mysqlCluster.Status.Repo == nil {
	cluster.Status.Repo = &apiv1alpha1.RepoStatus{}
	// }
	result := reconcile.Result{}
	repoHostName := cluster.GetName() + "-repo"
	repoHost := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoHostName,
			Namespace: cluster.GetNamespace(),
		},
		Spec: appsv1.StatefulSetSpec{
			// Selector: metav1.SetAsLabelSelector(cluster.GetObjectMeta().GetLabels()),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
			},
		},
	}
	// ensure conditions are set before returning as needed by subsequent reconcile functions
	defer func() {
		repoHostReady := metav1.Condition{
			ObservedGeneration: cluster.GetGeneration(),
			Type:               ConditionRepoHostReady,
		}
		if cluster.Status.Repo == nil {
			repoHostReady.Status = metav1.ConditionUnknown
			repoHostReady.Reason = "RepoHostStatusMissing"
			repoHostReady.Message = "MySQLsh dedicated repository host status is missing"
		} else if cluster.Status.Repo.Ready {
			repoHostReady.Status = metav1.ConditionTrue
			repoHostReady.Reason = "RepoHostReady"
			repoHostReady.Message = "MySQLsh dedicated repository host is ready"
		} else {
			repoHostReady.Status = metav1.ConditionFalse
			repoHostReady.Reason = "RepoHostNotReady"
			repoHostReady.Message = "MySQLsh dedicated repository host is not ready"
		}
		meta.SetStatusCondition(&cluster.Status.Repo.Conditions, repoHostReady)
	}()
	var isCreate bool

	if cluster.Spec.LogicalBackups.Enabled {
		isCreate = true
	}
	if !isCreate {
		return result, nil
	}
	fmt.Print(cluster.GetSelectorLabels())
	repoHost.Spec.Selector = metav1.SetAsLabelSelector(cluster.GetSelectorLabels())
	repoHost.Spec.Template.ObjectMeta.Labels = cluster.GetLabels()
	repoHost.Spec.Template.Spec.Containers = []corev1.Container{{
		Name:                     "mysqlsh",
		Image:                    cluster.Spec.LogicalBackups.Image,
		Command:                  []string{},
		Args:                     []string{},
		WorkingDir:               "",
		Ports:                    []corev1.ContainerPort{},
		EnvFrom:                  []corev1.EnvFromSource{},
		Env:                      []corev1.EnvVar{},
		Resources:                corev1.ResourceRequirements{},
		VolumeMounts:             []corev1.VolumeMount{{Name: "mysqlsh-repo", MountPath: "/repo"}},
		VolumeDevices:            []corev1.VolumeDevice{},
		LivenessProbe:            nil,
		ReadinessProbe:           nil,
		StartupProbe:             nil,
		Lifecycle:                &corev1.Lifecycle{},
		TerminationMessagePath:   "",
		TerminationMessagePolicy: "",
		ImagePullPolicy:          "",
		SecurityContext:          &corev1.SecurityContext{},
		Stdin:                    false,
		StdinOnce:                false,
		TTY:                      false,
	},
	}
	repoHost.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "mysqlsh-repo",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "mysqlsh-repo",
				}},
		},
	}
	repoHost.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysqlsh-repo",
				Namespace: cluster.GetNamespace(),
				Labels:    cluster.GetLabels(),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: cluster.Spec.LogicalBackups.Volume.VolumeClaimSpec.AccessModes,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: cluster.Spec.LogicalBackups.Volume.VolumeClaimSpec.Resources.Requests[corev1.ResourceStorage],
					},
				},
				StorageClassName: cluster.Spec.LogicalBackups.Volume.VolumeClaimSpec.StorageClassName,
			},
		},
	}
	err := errors.WithStack(r.apply(ctx, repoHost))
	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MySQLshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.MysqlCluster{}).
		Complete(r)
}

func (r *MySQLshReconciler) apply(ctx context.Context, object client.Object) error {
	// Generate an apply-patch by comparing the object to its zero value.
	zero := reflect.New(reflect.TypeOf(object).Elem()).Interface()
	data, err := client.MergeFrom(zero.(client.Object)).Data(object)
	apply := client.RawPatch(client.Apply.Type(), data)

	patch := utils.NewJSONPatch()

	// Send the apply-patch with force=true.
	if err == nil {
		err = r.patch(ctx, object, apply, client.ForceOwnership)
	}

	// Send the json-patch when necessary.
	if err == nil && !patch.IsEmpty() {
		err = r.patch(ctx, object, patch)
	}
	return err
}

// patch sends patch to object's endpoint in the Kubernetes API and updates
// object with any returned content. The fieldManager is set to r.Owner, but
// can be overridden in options.
// - https://docs.k8s.io/reference/using-api/server-side-apply/#managers
func (r *MySQLshReconciler) patch(
	ctx context.Context, object client.Object,
	patch client.Patch, options ...client.PatchOption,
) error {
	options = append([]client.PatchOption{r.Owner}, options...)
	return r.Client.Patch(ctx, object, patch, options...)
}
