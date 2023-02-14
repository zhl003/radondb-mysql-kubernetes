package v1beta1

import (
	v1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Backup{}

func (src *Backup) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.Backup)
	dst.Spec.ClusterName = src.Spec.ClusterName
	return nil
}

func (dst *Backup) ConvertFrom(dstRaw conversion.Hub) error {
	src := dstRaw.(*v1alpha1.Backup)
	dst.Spec.ClusterName = src.Spec.ClusterName
	return nil
}
