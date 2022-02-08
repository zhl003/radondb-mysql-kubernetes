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

package simplecase

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

// Simple and quick test case, used to verify that the E2E framework is available.
var _ = Describe("Namespace test", Label("simplecase"), Ordered, func() {
	var f *framework.Framework
	BeforeEach(func() {
		f = &framework.Framework{
			BaseName: "mysqlcluster-e2e",
			Log:      framework.Log,
		}
		f.BeforeEach()
	})

	It("test list namespace", Label("list ns"), func() {
		namespaces, err := f.ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		Expect(err).Should(BeNil())
		Expect(len(namespaces.Items)).ShouldNot(BeZero())
		for _, ns := range namespaces.Items {
			fmt.Fprintln(GinkgoWriter, ns.Name)
		}
	})
})
