package kube

import (
	"context"
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"github.com/sqc157400661/util"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcileContext declares the context for reconciliation.
type ReconcileContext interface {
	ReconcileHelper

	// Name is a helper method that returns the name of current reconcile object.
	Name() string
	// Namespace is a helper method that returns the namespace of current reconcile object.
	Namespace() string
	// Context returns the current reconcile context.
	Context() context.Context
	// Request returns the current reconcile request.
	Request() reconcile.Request
	// Owner returns the current reconcile fieldOwner.
	Owner() client.FieldOwner
	// Recorder returns the current reconcile recorder.
	Recorder() record.EventRecorder
	// Close closes the context to avoid resource leaks.

	Get(object client.Object) error
	List(list client.ObjectList, selector labels.Selector) error
	Patch(object client.Object, patch client.Patch, options ...client.PatchOption) error
	Apply(object client.Object) error                       // Server-Side Apply
	CSAApply(new client.Object, old ...client.Object) error // Client-Side Apply
	Close() error
}

type BaseReconcileContext struct {
	ReconcileHelper
	context  context.Context
	request  reconcile.Request
	owner    client.FieldOwner
	recorder record.EventRecorder
}

func (rc *BaseReconcileContext) Name() string {
	return rc.request.Name
}

func (rc *BaseReconcileContext) Namespace() string {
	return rc.request.Namespace
}
func (rc *BaseReconcileContext) Context() context.Context {
	return rc.context
}

func (rc *BaseReconcileContext) Request() reconcile.Request {
	return rc.request
}
func (rc *BaseReconcileContext) Owner() client.FieldOwner {
	return rc.owner
}

func (rc *BaseReconcileContext) Recorder() record.EventRecorder {
	return rc.recorder
}
func (rc *BaseReconcileContext) Close() error {
	return nil
}

func (rc *BaseReconcileContext) shallowCopy() *BaseReconcileContext {
	shallowCopy := *rc
	return &shallowCopy
}

func (rc *BaseReconcileContext) Get(object client.Object) error {
	return rc.Client().Get(rc.context, client.ObjectKeyFromObject(object), object)
}

func (rc *BaseReconcileContext) List(list client.ObjectList, selector labels.Selector) error {
	return rc.Client().List(rc.context, list,
		client.InNamespace(rc.Namespace()),
		client.MatchingLabelsSelector{Selector: selector},
	)
}

// Patch sends patch to object's endpoint in the Kubernetes API and updates
// object with any returned content. The fieldManager is set to r.Owner, but
// can be overridden in options.
// - https://docs.k8s.io/reference/using-api/server-side-apply/#managers
func (rc *BaseReconcileContext) Patch(object client.Object,
	patch client.Patch, options ...client.PatchOption,
) error {
	options = append([]client.PatchOption{rc.owner}, options...)
	return rc.Client().Patch(rc.context, object, patch, options...)
}

// when not support server-side apply, use it
func (rc *BaseReconcileContext) CSAApply(object client.Object, oldObject ...client.Object) error {
	var old client.Object
	var err error
	if len(oldObject) > 0 {
		old = oldObject[0]
		if old == nil {
			return client.IgnoreAlreadyExists(rc.Client().Create(rc.context, object))
		}
	} else {
		old = reflect.New(reflect.TypeOf(object).Elem()).Interface().(client.Object)
		err := rc.Client().Get(rc.context, client.ObjectKeyFromObject(object), old)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			return client.IgnoreAlreadyExists(rc.Client().Create(rc.context, object))
		}
	}
	patch := util.NewMergePatch()
	patch.PatchLabels(object.GetLabels(), old.GetLabels())
	patch.PatchAnnos(object.GetAnnotations(), old.GetAnnotations())
	switch actual := object.(type) {
	case *corev1.ConfigMap:
		oldObj := old.(*corev1.ConfigMap)
		patch.PatchStringMap([]string{"data"}, actual.Data, oldObj.Data)
	case *corev1.Endpoints:
		oldObj := old.(*corev1.Endpoints)
		if !equality.Semantic.DeepEqual(actual.Subsets, oldObj.Subsets) {
			patch.Add("subsets")(actual.Subsets)
		}
	case *corev1.PersistentVolumeClaim:
		oldObj := old.(*corev1.PersistentVolumeClaim)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *corev1.Secret:
		oldObj := old.(*corev1.Secret)
		patch.PatchByteMap([]string{"data"}, actual.Data, oldObj.Data)
	case *corev1.Service:
		oldObj := old.(*corev1.Service)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *corev1.ServiceAccount:
	case *appsv1.Deployment:
		oldObj := old.(*appsv1.Deployment)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *appsv1.StatefulSet:
		oldObj := old.(*appsv1.StatefulSet)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *batchv1.Job:
		oldObj := old.(*batchv1.Job)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *batchv1.CronJob:
		oldObj := old.(*batchv1.CronJob)
		if !equality.Semantic.DeepEqual(actual.Spec, oldObj.Spec) {
			patch.Add("spec")(actual.Spec)
		}
	case *rbacv1.RoleBinding:
		oldObj := old.(*rbacv1.RoleBinding)
		if !equality.Semantic.DeepEqual(actual.Subjects, oldObj.Subjects) {
			patch.Add("subjects")(actual.Subjects)
		}
	case *corev1.Pod:
		oldObj := old.(*corev1.Pod)
		if !equality.Semantic.DeepEqual(actual.Spec.Containers, oldObj.Spec.Containers) {
			patch.Add("spec")(actual.Spec)
		}
	default:
	}
	if !patch.IsEmpty() {
		fmt.Printf("type :%T do Patch\n", object)
		err = rc.Client().Patch(rc.context, object, patch)
	}
	return err
}

// apply sends an apply patch to object's endpoint in the Kubernetes API and
// updates object with any returned content. The fieldManager is set to
// r.Owner and the force parameter is true.
// - https://docs.k8s.io/reference/using-api/server-side-apply/#managers
// - https://docs.k8s.io/reference/using-api/server-side-apply/#conflicts
func (rc *BaseReconcileContext) Apply(object client.Object) error {
	// Generate an apply-patch by comparing the object to its zero value.
	zero := reflect.New(reflect.TypeOf(object).Elem()).Interface()
	data, err := client.MergeFrom(zero.(client.Object)).Data(object)
	apply := client.RawPatch(client.Apply.Type(), data)

	// Keep a copy of the object before any API calls.
	intent := object.DeepCopyObject()
	patch := util.NewJSONPatch()

	// Send the apply-patch with force=true.
	if err == nil {
		err = rc.Patch(object, apply, client.ForceOwnership)
	}

	// Some fields cannot be server-side applied correctly. When their outcome
	// does not match the intent, send a json-patch to get really specific.
	switch actual := object.(type) {
	case *corev1.Service:
		// Changing Service.Spec.Type requires a special apply-patch sometimes.
		if err != nil {
			err = rc.handleServiceError(object.(*corev1.Service), data, err)
		}

		applyServiceSpec(patch, actual.Spec, intent.(*corev1.Service).Spec, "spec")
	}

	// Send the json-patch when necessary.
	if err == nil && !patch.IsEmpty() {
		err = rc.Patch(object, patch)
	}
	return err
}

// handleServiceError inspects err for expected Kubernetes API responses to
// writing a Service. It returns err when it cannot resolve the issue, otherwise
// it returns nil.
func (rc *BaseReconcileContext) handleServiceError(
	service *corev1.Service, apply []byte, err error,
) error {
	var status metav1.Status
	if api := apierrors.APIStatus(nil); errors.As(err, &api) {
		status = api.Status()
	}

	// Service.Spec.Ports.NodePort must be cleared for ClusterIP prior to
	// Kubernetes 1.20. When all the errors are about disallowed "nodePort",
	// run a json-patch on the apply-patch to set them all to null.
	// - https://issue.k8s.io/33766
	if service.Spec.Type == corev1.ServiceTypeClusterIP {
		add := json.RawMessage(`"add"`)
		null := json.RawMessage(`null`)
		patch := make(jsonpatch.Patch, 0, len(service.Spec.Ports))

		if apierrors.IsInvalid(err) && status.Details != nil {
			for i, cause := range status.Details.Causes {
				path := json.RawMessage(fmt.Sprintf(`"/spec/ports/%d/nodePort"`, i))

				if cause.Type == metav1.CauseType(field.ErrorTypeForbidden) &&
					cause.Field == fmt.Sprintf("spec.ports[%d].nodePort", i) {
					patch = append(patch,
						jsonpatch.Operation{"op": &add, "value": &null, "path": &path})
				}
			}
		}

		// Amend the apply-patch when all the errors can be fixed.
		if len(patch) == len(service.Spec.Ports) {
			apply, err = patch.Apply(apply)
		}

		// Send the apply-patch with force=true.
		if err == nil {
			patch := client.RawPatch(client.Apply.Type(), apply)
			err = rc.Patch(service, patch, client.ForceOwnership)
		}
	}

	return err
}

// applyServiceSpec is called by Reconciler.apply to work around issues
// with server-side apply.
func applyServiceSpec(
	patch *util.JSON6902, actual, intent corev1.ServiceSpec, path ...string,
) {
	// Service.Spec.Selector is not +mapType=atomic until Kubernetes 1.22.
	// - https://issue.k8s.io/97970
	if !equality.Semantic.DeepEqual(actual.Selector, intent.Selector) {
		patch.Replace(append(path, "selector")...)(intent.Selector)
	}
}

func NewBaseReconcileContext(h ReconcileHelper, context context.Context, request reconcile.Request, owner client.FieldOwner, recorder record.EventRecorder) *BaseReconcileContext {
	return &BaseReconcileContext{
		ReconcileHelper: h,
		context:         context,
		request:         request,
		owner:           owner,
		recorder:        recorder,
	}
}
