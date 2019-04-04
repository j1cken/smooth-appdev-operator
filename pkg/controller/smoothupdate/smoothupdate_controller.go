package smoothupdate

import (
	"context"

	smoothopsv1alpha1 "github.com/j1cken/smooth-appdev-operator/pkg/apis/smoothops/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_smoothupdate")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SmoothUpdate Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSmoothUpdate{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("smoothupdate-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SmoothUpdate
	err = c.Watch(&source.Kind{Type: &smoothopsv1alpha1.SmoothUpdate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner SmoothUpdate
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &smoothopsv1alpha1.SmoothUpdate{},
	})

	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &smoothopsv1alpha1.SmoothUpdate{},
	// })
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSmoothUpdate{}

// ReconcileSmoothUpdate reconciles a SmoothUpdate object
type ReconcileSmoothUpdate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiservererr = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &smoothopsv1alpha1.SmoothUpdate{},

	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SmoothUpdate object and makes changes based on the state read
// and what is in the SmoothUpdate.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSmoothUpdate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SmoothUpdate")

	// Fetch the SmoothUpdate instance
	instance := &smoothopsv1alpha1.SmoothUpdate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("found version in CRD", "Update.version", instance.Spec.Version)

	mysqlSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "mysql", Namespace: instance.Namespace}, mysqlSecret)
	if err != nil {
		reqLogger.Error(err, "Failed to retrieve secret", "Deployment.Namespace", instance.Namespace, "Secret", "database-user")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	initialDeployment := false
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForSmoothUpdate(instance, mysqlSecret)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}
		initialDeployment = true
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the version is the same as the spec
	version := instance.Spec.Version

	if found.Spec.Template.ResourceVersion != version || initialDeployment == true {

		configmap := &corev1.ConfigMap{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "sql-updates", Namespace: instance.Namespace}, configmap)
		if err != nil {
			reqLogger.Error(err, "Failed to retrieve configmap", "Deployment.Namespace", instance.Namespace, "ConfigMap", "sql-updates")
			return reconcile.Result{}, err
		}

		pod := r.createSQLUpdatePod(instance, mysqlSecret, configmap)
		reqLogger.Info("mysql pod", "Pod", pod)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "Error while executing sql scripts")
			return reconcile.Result{}, err
		}

		if !initialDeployment {
			found.Spec.Template.ResourceVersion = version
			found.Spec.Template.Spec.Containers[0].Image = "docker-registry.default.svc:5000/" + instance.Namespace + "/" + instance.Spec.Deployment + ":" + instance.Spec.Version
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return reconcile.Result{}, err
			}
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSmoothUpdate) createSQLUpdatePod(m *smoothopsv1alpha1.SmoothUpdate, secret *corev1.Secret, configmap *corev1.ConfigMap) *corev1.Pod {
	database := string(secret.Data["database-name"])
	user := string(secret.Data["database-user"])
	password := string(secret.Data["database-password"])

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sql-updates-" + m.Spec.Version,
			Namespace: m.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "mysql-cli",
					Image:   "mysql",
					Command: []string{"mysql", "-h", "mysql", "-u", user, "-p" + password, "-D", database, "-e source /tmp/sql-updates/" + m.Spec.UpdateSQL + ";"},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "sql-update-configmap",
						MountPath: "/tmp/sql-updates",
					}},
				}},
			RestartPolicy: "Never",
			Volumes: []corev1.Volume{{
				Name: "sql-update-configmap",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "sql-updates",
						},
					},
				},
			}},
		}}
}

func (r *ReconcileSmoothUpdate) deploymentForSmoothUpdate(m *smoothopsv1alpha1.SmoothUpdate, secret *corev1.Secret) *appsv1.Deployment {
	ls := labelsForSmoothUpdate(m.Name, m.Spec.Deployment)
	database := string(secret.Data["database-name"])
	user := string(secret.Data["database-user"])
	password := string(secret.Data["database-password"])

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:          ls,
					ResourceVersion: m.Spec.Version,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "docker-registry.default.svc:5000/" + m.Namespace + "/" + m.Spec.Deployment + ":" + m.Spec.Version,
						Name:  m.Spec.Deployment,
						// exec java -XX:+UseParallelOldGC -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:MinHeapFreeRatio=10 -XX:MaxHeapFreeRatio=20 -XX:GCTimeRatio=4 -XX:AdaptiveSizePolicyWeight=90 -XX:MaxMetaspaceSize=100m -XX:+ExitOnOutOfMemoryError -cp . -jar /deployments/quarkus-microprofile-rest-1.0-SNAPSHOT-runner.jar
						// !!!!!!!!!!!!!!!!!! 1.0-SNAPSHOT needs to be removed !!!!!!!!!!!!!!!!!!
						Command: []string{"java", "-Dquarkus.datasource.url=jdbc:mysql://mysql." + m.Namespace + ".svc.cluster.local/" + database, "-Dquarkus.datasource.username=" + user, "-Dquarkus.datasource.password=" + password, "-XX:+UseParallelOldGC", "-XX:+UnlockExperimentalVMOptions", "-XX:+UseCGroupMemoryLimitForHeap", "-XX:MinHeapFreeRatio=10", "-XX:MaxHeapFreeRatio=20", "-XX:GCTimeRatio=4", "-XX:AdaptiveSizePolicyWeight=90", "-XX:MaxMetaspaceSize=100m", "-XX:+ExitOnOutOfMemoryError", "-cp", ".", "-jar", "/deployments/quarkus-microprofile-rest-1.0-SNAPSHOT-runner.jar"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						},
							{
								ContainerPort: 8443,
								Name:          "https",
							},
							{
								ContainerPort: 8778,
								Name:          "jolokia",
							},
						},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForSmoothUpdate(name string, app string) map[string]string {
	return map[string]string{"app": app, "update_cr": name}
}
