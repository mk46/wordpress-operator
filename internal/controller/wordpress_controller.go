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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cachev1alpha1 "github.com/mk46/wordpress-operator/api/v1alpha1"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.example.com,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=wordpresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=wordpresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wordpress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *WordpressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Wordpress instance
	var wordpress cachev1alpha1.Wordpress
	if err := r.Get(ctx, req.NamespacedName, &wordpress); err != nil {
		log.Error(err, "unable to fetch Wordpress")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define the desired state of the ConfigMap

	// Fetch data from config
	configMapData := make(map[string]string)
	for k, v := range wordpress.Spec.ConfigData {
		configMapData[k] = v
	}

	// Create ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name + "-config",
			Namespace: wordpress.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&wordpress, schema.GroupVersionKind{
					Group:   cachev1alpha1.GroupVersion.Group,
					Version: cachev1alpha1.GroupVersion.Version,
					Kind:    "Wordpress",
				}),
			},
		},
		Data: configMapData,
	}

	// Create or Update the ConfigMap
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data = configMapData

		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update ConfigMap")
		return ctrl.Result{}, err
	}
	log.Info("ConfigMap reconciled", "operation", op, "name", configMap.Name)

	// Define the desired state of the Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name + "-secret",
			Namespace: wordpress.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&wordpress, schema.GroupVersionKind{
					Group:   cachev1alpha1.GroupVersion.Group,
					Version: cachev1alpha1.GroupVersion.Version,
					Kind:    "Wordpress",
				}),
			},
		},
		StringData: map[string]string{
			"dbUsername": wordpress.Spec.DBUserName,
			"dbPassword": wordpress.Spec.DBPassword,
		},
	}

	// Create or Update the Secret
	op, err = controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.StringData == nil {
			secret.StringData = make(map[string]string)
		}
		secret.StringData["dbUsername"] = wordpress.Spec.DBUserName
		secret.StringData["dbPassword"] = wordpress.Spec.DBPassword
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update Secret")
		return ctrl.Result{}, err
	}
	log.Info("Secret reconciled", "operation", op, "name", secret.Name)

	// Define the desired state of the Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name,
			Namespace: wordpress.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&wordpress, schema.GroupVersionKind{
					Group:   cachev1alpha1.GroupVersion.Group,
					Version: cachev1alpha1.GroupVersion.Version,
					Kind:    "Wordpress",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &wordpress.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": wordpress.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": wordpress.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "wordpress",
						Image: wordpress.Spec.Image,
						// Fetched env data from Secret
						Env: []corev1.EnvVar{
							{
								Name: "WORDPRESS_DB_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: wordpress.Name + "-secret",
										},
										Key: "dbUsername",
									},
								},
							},
							{
								Name: "WORDPRESS_DB_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: wordpress.Name + "-secret",
										},
										Key: "dbPassword",
									},
								},
							},
						},

						// Fetched env data from ConfigMap
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: wordpress.Name + "-config",
									},
								},
							},
						},
					}},
				},
			},
		},
	}

	// Create or Update the Deployment
	op, err = controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		deployment.Spec.Replicas = &wordpress.Spec.Replicas
		deployment.Spec.Template.Spec.Containers[0].Image = wordpress.Spec.Image
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update Deployment")
		return ctrl.Result{}, err
	}
	log.Info("Deployment reconciled", "operation", op, "name", deployment.Name)

	// Define the desired state of the HPA
	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wordpress.Name,
			Namespace: wordpress.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&wordpress, schema.GroupVersionKind{
					Group:   cachev1alpha1.GroupVersion.Group,
					Version: cachev1alpha1.GroupVersion.Version,
					Kind:    "Wordpress",
				}),
			},
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       wordpress.Name,
				APIVersion: "apps/v1",
			},
			MinReplicas:                    &wordpress.Spec.MinReplicas,
			MaxReplicas:                    wordpress.Spec.MaxReplicas,
			TargetCPUUtilizationPercentage: &wordpress.Spec.TargetCPUUtilizationPercentage,
		},
	}

	// Create or Update the HPA
	op, err = controllerutil.CreateOrUpdate(ctx, r.Client, hpa, func() error {
		hpa.Spec.MinReplicas = &wordpress.Spec.MinReplicas
		hpa.Spec.MaxReplicas = wordpress.Spec.MaxReplicas
		hpa.Spec.TargetCPUUtilizationPercentage = &wordpress.Spec.TargetCPUUtilizationPercentage
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update HPA")
		return ctrl.Result{}, err
	}
	log.Info("HPA reconciled", "operation", op, "name", hpa.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Wordpress{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
