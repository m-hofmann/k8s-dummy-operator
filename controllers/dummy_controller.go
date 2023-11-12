/*
Copyright 2023.

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
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	interviewcomv1alpha1 "github.com/m-hofmann/k8s-dummy-operator/api/v1alpha1"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;create;delete;update;watch;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	dummy := &interviewcomv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, dummy)
	if err != nil {
		if errors.IsNotFound(err) {
			// Dummy not found, has probably been deleted
			// do not requeue
			logger.Info("Dummy not found. Ignoring deleted object")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to retrieve Dummy object")
		return ctrl.Result{}, err
	}

	logger.Info("Processing Dummy object", "name", req.Name, "namespace", req.Namespace, "message", dummy.Spec.Message)

	result, err := r.reconcilePod(dummy, logger)
	if err != nil {
		logger.Error(err, "Failed to reconcile Dummy status")
		return result, err
	}

	result, err = r.updateStatusMessage(dummy, err, logger)
	if err != nil {
		logger.Error(err, "Failed to update Dummy status")
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *DummyReconciler) updateStatusMessage(dummy *interviewcomv1alpha1.Dummy, err error, logger logr.Logger) (ctrl.Result, error) {
	dummy.Status.SpecEcho = dummy.Spec.Message
	dummy, err = r.updateDummyStatus(dummy, logger)
	return ctrl.Result{}, err
}

func (r *DummyReconciler) updateDummyStatus(cr *interviewcomv1alpha1.Dummy, log logr.Logger) (*interviewcomv1alpha1.Dummy, error) {
	dummy := &interviewcomv1alpha1.Dummy{}
	err := r.Get(context.Background(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, dummy)
	if err != nil {
		return dummy, err
	}

	if !reflect.DeepEqual(cr.Status, dummy.Status) {
		log.Info("Updating Dummy status.")
		err = r.Status().Update(context.Background(), cr)
		if err != nil {
			return cr, err
		}
		updatedDummy := &interviewcomv1alpha1.Dummy{}
		err = r.Get(context.Background(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, updatedDummy)
		if err != nil {
			return cr, err
		}
		cr = updatedDummy.DeepCopy()
	}

	return cr, nil
}

func (r *DummyReconciler) reconcilePod(dummy *interviewcomv1alpha1.Dummy, logger logr.Logger) (ctrl.Result, error) {
	pod := newPodForDummy(dummy)
	if err := ctrl.SetControllerReference(dummy, pod, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	existingPod := &v1.Pod{}
	// create pod as needed
	err := r.Get(context.Background(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, existingPod)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating new Pod.", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(context.Background(), pod)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		logger.Info("Pod already exists.", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	}

	// retrieve pod status
	dummy.Status.PodStatus = existingPod.Status.Phase
	logger.Info("Reflecting Pod status to Dummy object.", "Pod.Status", dummy.Status.PodStatus)
	dummy, err = r.updateDummyStatus(dummy, logger)
	if err != nil {
		logger.Error(err, "Failed to update Dummy status.")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func newPodForDummy(dummy *interviewcomv1alpha1.Dummy) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummy.Name,
			Namespace: dummy.Namespace,
			Labels: map[string]string{
				"dummy": dummy.Name,
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "nginx:latest",
					Name:  "nginx",
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewcomv1alpha1.Dummy{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
