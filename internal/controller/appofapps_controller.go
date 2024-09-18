/*
Copyright 2024.

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

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	visualizationv1alpha1 "github.com/maksimstankevic/app-of-apps-visualizer/api/v1alpha1"
)

// Definitions to manage status conditions
const (
	// typeAvailableAppOfApps represents the status of the AppVersion reconciliation
	typeAvailableAppOfApps = "Available"
	// typeDegradedAppOfApps represents the status used when the custom resource is deleted and the finalizer operations are yet to occur.
	// typeDegradedAppOfApps = "Degraded"
)

// AppOfAppsReconciler reconciles a AppOfApps object
type AppOfAppsReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appofapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appofapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appofapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=visualization.magiccicd,resources=appversions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppOfApps object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *AppOfAppsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Memcached instance
	// The purpose is check if the Custom Resource for the Kind Memcached
	// is applied on the cluster if not we return nil to stop the reconciliation
	appOfApps := &visualizationv1alpha1.AppOfApps{}
	err := r.Get(ctx, req.NamespacedName, appOfApps)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("appofapps resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get appofapps")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status is available
	if len(appOfApps.Status.Conditions) == 0 {
		log.Info("In status update!!!!!!!!!!!!!!!!!!!!!")
		meta.SetStatusCondition(&appOfApps.Status.Conditions, metav1.Condition{Type: typeAvailableAppOfApps, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, appOfApps); err != nil {
			log.Error(err, "Failed to update AppOfApps status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the memcached Custom Resource after updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, appOfApps); err != nil {
			log.Info("In refetchung!!!!!!!!!!!!!!!!!!!!!")
			log.Error(err, "Failed to re-fetch AppOfApps")
			return ctrl.Result{}, err
		}
	}

	// // Let's add a finalizer. Then, we can define some operations which should
	// // occur before the custom resource is deleted.
	// // More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	// if !controllerutil.ContainsFinalizer(memcached, memcachedFinalizer) {
	// 	log.Info("Adding Finalizer for Memcached")
	// 	if ok := controllerutil.AddFinalizer(memcached, memcachedFinalizer); !ok {
	// 		log.Error(err, "Failed to add finalizer into the custom resource")
	// 		return ctrl.Result{Requeue: true}, nil
	// 	}

	// 	if err = r.Update(ctx, memcached); err != nil {
	// 		log.Error(err, "Failed to update custom resource to add finalizer")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// // Check if the Memcached instance is marked to be deleted, which is
	// // indicated by the deletion timestamp being set.
	// isMemcachedMarkedToBeDeleted := memcached.GetDeletionTimestamp() != nil
	// if isMemcachedMarkedToBeDeleted {
	// 	if controllerutil.ContainsFinalizer(memcached, memcachedFinalizer) {
	// 		log.Info("Performing Finalizer Operations for Memcached before delete CR")

	// 		// Let's add here a status "Downgrade" to reflect that this resource began its process to be terminated.
	// 		meta.SetStatusCondition(&memcached.Status.Conditions, metav1.Condition{Type: typeDegradedMemcached,
	// 			Status: metav1.ConditionUnknown, Reason: "Finalizing",
	// 			Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", memcached.Name)})

	// 		if err := r.Status().Update(ctx, memcached); err != nil {
	// 			log.Error(err, "Failed to update Memcached status")
	// 			return ctrl.Result{}, err
	// 		}

	// 		// Perform all operations required before removing the finalizer and allow
	// 		// the Kubernetes API to remove the custom resource.
	// 		r.doFinalizerOperationsForMemcached(memcached)

	// 		// TODO(user): If you add operations to the doFinalizerOperationsForMemcached method
	// 		// then you need to ensure that all worked fine before deleting and updating the Downgrade status
	// 		// otherwise, you should requeue here.

	// 		// Re-fetch the memcached Custom Resource before updating the status
	// 		// so that we have the latest state of the resource on the cluster and we will avoid
	// 		// raising the error "the object has been modified, please apply
	// 		// your changes to the latest version and try again" which would re-trigger the reconciliation
	// 		if err := r.Get(ctx, req.NamespacedName, memcached); err != nil {
	// 			log.Error(err, "Failed to re-fetch memcached")
	// 			return ctrl.Result{}, err
	// 		}

	// 		meta.SetStatusCondition(&memcached.Status.Conditions, metav1.Condition{Type: typeDegradedMemcached,
	// 			Status: metav1.ConditionTrue, Reason: "Finalizing",
	// 			Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", memcached.Name)})

	// 		if err := r.Status().Update(ctx, memcached); err != nil {
	// 			log.Error(err, "Failed to update Memcached status")
	// 			return ctrl.Result{}, err
	// 		}

	// 		log.Info("Removing Finalizer for Memcached after successfully perform the operations")
	// 		if ok := controllerutil.RemoveFinalizer(memcached, memcachedFinalizer); !ok {
	// 			log.Error(err, "Failed to remove finalizer for Memcached")
	// 			return ctrl.Result{Requeue: true}, nil
	// 		}

	// 		if err := r.Update(ctx, memcached); err != nil {
	// 			log.Error(err, "Failed to remove finalizer for Memcached")
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// 	return ctrl.Result{}, nil
	// }

	// // Check if the deployment already exists, if not create a new one
	// found := &appsv1.Deployment{}
	// err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)
	// if err != nil && apierrors.IsNotFound(err) {
	// 	// Define a new deployment
	// 	dep, err := r.deploymentForMemcached(memcached)
	// 	if err != nil {
	// 		log.Error(err, "Failed to define new Deployment resource for Memcached")

	// 		// The following implementation will update the status
	// 		meta.SetStatusCondition(&memcached.Status.Conditions, metav1.Condition{Type: typeAvailableMemcached,
	// 			Status: metav1.ConditionFalse, Reason: "Reconciling",
	// 			Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", memcached.Name, err)})

	// 		if err := r.Status().Update(ctx, memcached); err != nil {
	// 			log.Error(err, "Failed to update Memcached status")
	// 			return ctrl.Result{}, err
	// 		}

	// 		return ctrl.Result{}, err
	// 	}

	// 	log.Info("Creating a new Deployment",
	// 		"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 	if err = r.Create(ctx, dep); err != nil {
	// 		log.Error(err, "Failed to create new Deployment",
	// 			"Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 		return ctrl.Result{}, err
	// 	}

	// 	// Deployment created successfully
	// 	// We will requeue the reconciliation so that we can ensure the state
	// 	// and move forward for the next operations
	// 	return ctrl.Result{RequeueAfter: time.Minute}, nil
	// } else if err != nil {
	// 	log.Error(err, "Failed to get Deployment")
	// 	// Let's return the error for the reconciliation be re-trigged again
	// 	return ctrl.Result{}, err
	// }

	// // The CRD API defines that the Memcached type have a MemcachedSpec.Size field
	// // to set the quantity of Deployment instances to the desired state on the cluster.
	// // Therefore, the following code will ensure the Deployment size is the same as defined
	// // via the Size spec of the Custom Resource which we are reconciling.
	// size := memcached.Spec.Size
	// if *found.Spec.Replicas != size {
	// 	found.Spec.Replicas = &size
	// 	if err = r.Update(ctx, found); err != nil {
	// 		log.Error(err, "Failed to update Deployment",
	// 			"Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

	// 		// Re-fetch the memcached Custom Resource before updating the status
	// 		// so that we have the latest state of the resource on the cluster and we will avoid
	// 		// raising the error "the object has been modified, please apply
	// 		// your changes to the latest version and try again" which would re-trigger the reconciliation
	// 		if err := r.Get(ctx, req.NamespacedName, memcached); err != nil {
	// 			log.Error(err, "Failed to re-fetch memcached")
	// 			return ctrl.Result{}, err
	// 		}

	// 		// The following implementation will update the status
	// 		meta.SetStatusCondition(&memcached.Status.Conditions, metav1.Condition{Type: typeAvailableMemcached,
	// 			Status: metav1.ConditionFalse, Reason: "Resizing",
	// 			Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", memcached.Name, err)})

	// 		if err := r.Status().Update(ctx, memcached); err != nil {
	// 			log.Error(err, "Failed to update Memcached status")
	// 			return ctrl.Result{}, err
	// 		}

	// 		return ctrl.Result{}, err
	// 	}

	// 	// Now, that we update the size we want to requeue the reconciliation
	// 	// so that we can ensure that we have the latest state of the resource before
	// 	// update. Also, it will help ensure the desired state on the cluster
	// 	return ctrl.Result{Requeue: true}, nil
	// }

	// // The following implementation will update the status
	// meta.SetStatusCondition(&memcached.Status.Conditions, metav1.Condition{Type: typeAvailableMemcached,
	// 	Status: metav1.ConditionTrue, Reason: "Reconciling",
	// 	Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", memcached.Name, size)})

	// if err := r.Status().Update(ctx, memcached); err != nil {
	// 	log.Error(err, "Failed to update Memcached status")
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppOfAppsReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&visualizationv1alpha1.AppOfApps{}).
		Complete(r)
}
