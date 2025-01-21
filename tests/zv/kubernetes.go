// Copyright 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zv

import (
	"context"
	"fmt"
	"strings"

	"github.com/openebs/lib-csi/pkg/common/errors"
	apis "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	client "github.com/openebs/lib-csi/pkg/common/kubernetes/client"
	"k8s.io/client-go/kubernetes"
)

// getClientsetFn is a typed function that
// abstracts fetching of clientset
type getClientsetFn func() (clientset *kubernetes.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(kubeConfigPath string) (clientset *kubernetes.Clientset, err error)

// getpvcFn is a typed function that
// abstracts fetching of pvc
type getFn func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.ZFSVolume, error)

// listFn is a typed function that abstracts
// listing of pvcs
type listFn func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error)

// deleteFn is a typed function that abstracts
// deletion of pvcs
type deleteFn func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error

// deleteFn is a typed function that abstracts
// deletion of pvc's collection
type deleteCollectionFn func(cli *kubernetes.Clientset, namespace string, listOpts metav1.ListOptions, deleteOpts *metav1.DeleteOptions) error

// createFn is a typed function that abstracts
// creation of pvc
type createFn func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)

// updateFn is a typed function that abstracts
// updation of pvc
type updateFn func(cli *kubernetes.Clientset, namespace string, pvc *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error)

// Kubeclient enables kubernetes API operations
// on pvc instance
type Kubeclient struct {
	// clientset refers to pvc clientset
	// that will be responsible to
	// make kubernetes API calls
	clientset *kubernetes.Clientset

	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// kubeconfig path to get kubernetes clientset
	kubeConfigPath string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	list                listFn
	get                 getFn
	// create        createFn
	// update        updateFn
	del deleteFn
	// delCollection deleteCollectionFn
}

// KubeclientBuildOption abstracts creating an
// instance of kubeclient
type KubeclientBuildOption func(*Kubeclient)

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = func() (clients *kubernetes.Clientset, err error) {
			return client.New().Clientset()
		}
	}

	if k.getClientsetForPath == nil {
		k.getClientsetForPath = func(kubeConfigPath string) (clients *kubernetes.Clientset, err error) {
			return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
		}
	}

	if k.get == nil {
		k.get = func(cli *kubernetes.Clientset, name string, namespace string, opts metav1.GetOptions) (*apis.ZFSVolume, error) {
			fmt.Println("YAHHh$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$####################################################################")
			//return cli.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, opts)
			//return client.New(client.WithKubeConfigPath(kubeConfigPath)).Clientset()
			//return ZfsV1Interface.ZFSVolumesGetter(namespace).Get(context.TODO(), name, opts)
			return &apis.ZFSVolume{}, nil
		}
	}

	if k.list == nil {
		k.list = func(cli *kubernetes.Clientset, namespace string, opts metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
			return cli.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), opts)
		}
	}

	if k.del == nil {
		k.del = func(cli *kubernetes.Clientset, namespace string, name string, deleteOpts *metav1.DeleteOptions) error {
			return cli.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, *deleteOpts)
		}
	}

}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *kubernetes.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithKubeConfigPath sets the kubeConfig path
// against client instance
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeClient returns a new instance of kubeclient meant for
// zv operations
func NewKubeClient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (*kubernetes.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientsetOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientsetOrCached() (*kubernetes.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}

	cs, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clientset")
	}
	k.clientset = cs
	return k.clientset, nil
}

// Get returns a pvc resource
// instances present in kubernetes cluster
func (k *Kubeclient) Get(name string, opts metav1.GetOptions) (*apis.ZFSVolume, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("failed to get zv: missing zv name")
	}
	cli, err := k.getClientsetOrCached()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get zv {%s}", name)
	}
	return k.get(cli, name, k.namespace, opts)
}
