package k8s

import (
	"bytes"
	"errors"
	"github.com/sqc157400661/util"
	"io"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExecOptions struct {
	Logger  *logr.Logger
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Timeout time.Duration
}

func (opts *ExecOptions) setDefaults() {
	// Set default timeout to 5s if not specified.
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}

	// Branch stdin/stdout/stderr are all null, set one.
	if opts.Stdin == nil && opts.Stdout == nil && opts.Stderr == nil {
		opts.Stdout = &bytes.Buffer{}
	}
}

type ReconcilePodExecCommandHelper interface {
	PodExec(pod *corev1.Pod, container string, command []string, opts ExecOptions) error
}

// ReconcileHelper declares the reconcile helper.
type ReconcileHelper interface {
	ReconcilePodExecCommandHelper

	Debug() bool

	ForceRequeueAfter() time.Duration
	ResetForceRequeueAfter(d time.Duration)

	// Client returns a API client of k8s.
	Client() client.Client
	// RestConfig returns the rest config used by client.
	RestConfig() *rest.Config
	// ClientSet returns the client set.
	ClientSet() *kubernetes.Clientset
	// Scheme returns the currently using scheme.
	Scheme() *runtime.Scheme
}

type DefaultReconcileHelper struct {
	client     client.Client
	restConfig *rest.Config
	clientSet  *kubernetes.Clientset
	scheme     *runtime.Scheme

	forceRequeueAfter time.Duration
}

func (rc *DefaultReconcileHelper) Debug() bool {
	return false
}

func (rc *DefaultReconcileHelper) PodExec(pod *corev1.Pod, container string, command []string, opts ExecOptions) error {
	if util.GetContainerFromPod(pod, container) == nil {
		return errors.New("container " + container + " not found in pod " + pod.Name)
	}

	// Set defaults.
	opts.setDefaults()

	logger := opts.Logger
	if logger != nil {
		logger.Info("Executing command", "pod", pod.Name, "container", container, "command", command, "timeout", opts.Timeout)
	}

	// Start execute
	req := rc.clientSet.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Timeout(opts.Timeout)
	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		//TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(rc.restConfig, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
	})
	return err
}

func (rc *DefaultReconcileHelper) Client() client.Client {
	return rc.client
}

func (rc *DefaultReconcileHelper) RestConfig() *rest.Config {
	return rc.restConfig
}

func (rc *DefaultReconcileHelper) ClientSet() *kubernetes.Clientset {
	return rc.clientSet
}

func (rc *DefaultReconcileHelper) Scheme() *runtime.Scheme {
	return rc.scheme
}

func (rc *DefaultReconcileHelper) ForceRequeueAfter() time.Duration {
	return rc.forceRequeueAfter
}

func (rc *DefaultReconcileHelper) ResetForceRequeueAfter(d time.Duration) {
	rc.forceRequeueAfter = d
}

func NewDefaultReconcileHelper(
	client client.Client,
	restConfig *rest.Config,
	clientSet *kubernetes.Clientset,
	scheme *runtime.Scheme,
) *DefaultReconcileHelper {
	return &DefaultReconcileHelper{
		client:     client,
		restConfig: restConfig,
		clientSet:  clientSet,
		scheme:     scheme,
	}
}

func NewDefaultReconcileHelperWithManager(mgr manager.Manager) (helper *DefaultReconcileHelper, err error) {
	clientSet, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return
	}
	helper = &DefaultReconcileHelper{
		client:     mgr.GetClient(),
		restConfig: mgr.GetConfig(),
		clientSet:  clientSet,
		scheme:     mgr.GetScheme(),
	}
	return
}
