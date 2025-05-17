package ride

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/n3wscott/theme-park-provider/api/v1alpha1"
)

// ConnectorWrapper wraps the connector for gRPC support.
type ConnectorWrapper struct {
	Log logging.Logger
}

// Connect implements the TypedExternalConnector interface.
func (c *ConnectorWrapper) Connect(ctx context.Context, mg resource.Managed) (managed.TypedExternalClient[resource.Managed], error) {
	log := c.Log
	if log == nil {
		log = logging.NewNopLogger()
	}
	conn := &connector{log: log}
	return conn.Connect(ctx, mg)
}

// connector satisfies the resource.ExternalConnector interface.
type connector struct {
	log logging.Logger
}

// Connect to the supplied resource.Managed (presumed to be a Ride) by using the Provider.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	c.log.Debug("Connecting to provider")

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return nil, errors.New("managed resource is not a Ride")
	}

	i.Status.SetConditions(Connecting())

	return &external{log: c.log}, nil
}

const TypeOperational xpv1.ConditionType = "Operational"

func Connecting() xpv1.Condition {
	return xpv1.Condition{
		Type:               TypeOperational,
		Status:             corev1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
		Reason:             xpv1.ReasonCreating,
	}
}

func Operating() xpv1.Condition {
	return xpv1.Condition{
		Type:               TypeOperational,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "Operating",
	}
}

func ShortStaffed() xpv1.Condition {
	return xpv1.Condition{
		Type:               TypeOperational,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             "ShortStaffed",
	}
}

// External satisfies the resource.ExternalClient interface.
type external struct {
	log logging.Logger
}

// Observe the existing external resource, if any. The managed.Reconciler
// calls Observe in order to determine whether an external resource needs to be
// created, updated, or deleted.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	gvk := mg.GetObjectKind().GroupVersionKind().String()
	e.log.Debug("Observing", "type", gvk)

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
	gvk := mg.GetObjectKind().GroupVersionKind().String()
	e.log.Debug("Create", "type", gvk)

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
	gvk := mg.GetObjectKind().GroupVersionKind().String()
	e.log.Debug("Update", "type", gvk)

	i, ok := mg.(*v1alpha1.Ride)
	if !ok {
		return managed.ExternalUpdate{}, errors.New("managed resource is not a Ride")
	}

	i.Status.RidersPerHour = 0

	ros := new(v1alpha1.RideOperatorList)

	//if err := e.client.List(ctx, ros); err != nil {
	//	return managed.ExternalUpdate{}, err
	//}

	var ro *v1alpha1.RideOperator
	// Look for an operator for this ride.
	for _, x := range ros.Items {
		if x.Spec.ForProvider.Ride.Name == i.GetName() {
			ro = &x
		}
	}

	if ro != nil {
		i.SetConditions(Operating())
		i.Status.Operator = &xpv1.TypedReference{
			APIVersion: ro.APIVersion,
			Kind:       ro.Kind,
			Name:       ro.Name,
			UID:        ro.UID,
		}
		i.Status.RidersPerHour = i.Spec.ForProvider.Capacity * ro.Spec.ForProvider.Frequency
	} else {
		i.SetConditions(ShortStaffed())
		i.Status.RidersPerHour = 0
	}

	return managed.ExternalUpdate{}, nil
}

// Delete the external resource. managed.Reconciler only calls Delete
// when a managed resource with the 'Delete' deletion policy (the default) has
// been deleted.
func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	gvk := mg.GetObjectKind().GroupVersionKind().String()
	e.log.Debug("Delete", "type", gvk)

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
