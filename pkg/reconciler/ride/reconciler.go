package ride

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/n3wscott/theme-park-provider/api/v1alpha1"
)

type Controller struct{}

// SetupWithManager instantiates a new controller using a managed.Reconciler configured to reconcile Ride.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(strings.ToLower(fmt.Sprintf("%s.%s", v1alpha1.RideKind, v1alpha1.GroupVersion.Group))).
		For(&v1alpha1.Ride{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RideGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()})))
}

// Connecter satisfies the resource.ExternalConnector interface.
type connecter struct{ client client.Client }

// Connect to the supplied resource.Managed (presumed to be a Ride) by using the Provider.
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Connecting to provider")
	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return nil, errors.New("managed resource is not a Ride")
	}

	i.Status.SetConditions(Connecting())

	return &external{client: c.client}, nil
}

func Connecting() xpv1.Condition {
	return xpv1.Condition{
		Type:               xpv1.TypeReady,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             xpv1.ReasonCreating,
	}
}

// External satisfies the resource.ExternalClient interface.
type external struct{ client client.Client }

// Observe the existing external resource, if any. The managed.Reconciler
// calls Observe in order to determine whether an external resource needs to be
// created, updated, or deleted.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Observing", "type", mg.GetObjectKind().GroupVersionKind().String())

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return managed.ExternalObservation{}, errors.New("managed resource is not a Ride")
	}

	_ = i

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: false,
		ConnectionDetails: managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretUserKey:     []byte("user"),
			xpv1.ResourceCredentialsSecretEndpointKey: []byte("host"),
		},
	}

	return o, nil
}

// Create a new external resource based on the specification of our managed
// resource. managed.Reconciler only calls Create if Observe reported
// that the external resource did not exist.
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Create", "type", mg.GetObjectKind().GroupVersionKind().String())

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return managed.ExternalCreation{}, errors.New("managed resource is not a Ride")
	}
	// Indicate that we're about to create the instance. Remember ExternalClient
	// authors can use a bespoke condition reason here in cases where Creating
	// doesn't make sense.
	i.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{ConnectionDetails: map[string][]byte{"ride": []byte("maybe")}}, nil
}

// Update the existing external resource to match the specifications of our
// managed resource. managed.Reconciler only calls Update if Observe
// reported that the external resource was not up to date.
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Update", "type", mg.GetObjectKind().GroupVersionKind().String())

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return managed.ExternalUpdate{}, errors.New("managed resource is not a Ride")
	}

	i.Status.RidersPerHour = 0

	ros := new(v1alpha1.RideOperatorList)

	if err := e.client.List(ctx, ros); err != nil {
		return managed.ExternalUpdate{}, err
	}

	var ro *v1alpha1.RideOperator
	// Look for an operator for this ride.
	for _, x := range ros.Items {
		if x.Spec.ForProvider.Ride.Name == i.GetName() {
			ro = &x
		}
	}

	if ro != nil {
		i.SetConditions(xpv1.Available())
		i.Status.Operator = &xpv1.TypedReference{
			APIVersion: ro.APIVersion,
			Kind:       ro.Kind,
			Name:       ro.Name,
			UID:        ro.UID,
		}
		i.Status.RidersPerHour = i.Spec.ForProvider.Capacity * ro.Spec.ForProvider.Frequency
	} else {
		i.SetConditions(xpv1.Unavailable())
		i.Status.RidersPerHour = 0
	}

	return managed.ExternalUpdate{}, nil
}

// Delete the external resource. managed.Reconciler only calls Delete
// when a managed resource with the 'Delete' deletion policy (the default) has
// been deleted.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Delete", "type", mg.GetObjectKind().GroupVersionKind().String())

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return managed.ExternalDelete{}, errors.New("managed resource is not a Ride")
	}
	// Indicate that we're about to delete the instance.
	i.SetConditions(xpv1.Deleting())

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
