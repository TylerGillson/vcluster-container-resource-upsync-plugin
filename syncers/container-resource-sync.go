package syncers

import (
	"fmt"

	"github.com/TylerGillson/vcluster-sdk/syncer"
	synccontext "github.com/TylerGillson/vcluster-sdk/syncer/context"
	"github.com/TylerGillson/vcluster-sdk/syncer/translator"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewContainerResourceSyncer(ctx *synccontext.RegisterContext) syncer.Syncer {
	return &containerResourceSyncer{
		NamespacedTranslator: translator.NewNamespacedTranslator(ctx, "pod", &corev1.Pod{}),
	}
}

type containerResourceSyncer struct {
	// implicitly uses default PhysicalToVirtual & VirtualToPhysical implementations
	translator.NamespacedTranslator
}

func (s *containerResourceSyncer) Name() string {
	return "container-resource-syncer"
}

func (s *containerResourceSyncer) Resource() client.Object {
	return &corev1.Pod{}
}

var _ syncer.Starter = &containerResourceSyncer{}

func (s *containerResourceSyncer) ReconcileStart(ctx *synccontext.SyncContext, req ctrl.Request) (bool, error) {
	return false, nil
}

func (s *containerResourceSyncer) ReconcileEnd() {
	// NOOP
}

func (s *containerResourceSyncer) Sync(ctx *synccontext.SyncContext, pObj client.Object, vObj client.Object) (ctrl.Result, error) {
	pPod := pObj.(*corev1.Pod)
	vPod := vObj.(*corev1.Pod)

	updated := s.updateContainerResources(pPod, vPod)
	if updated == nil {
		// no update is needed
		return ctrl.Result{}, nil
	}

	err := ctx.VirtualClient.Update(ctx.Context, updated)
	if err == nil {
		ctx.Log.Infof("updated pod %s/%s", vObj.GetNamespace(), vObj.GetName())
	} else {
		err = fmt.Errorf("failed to update pod %s/%s: %v", vObj.GetNamespace(), vObj.GetName(), err)
	}

	return ctrl.Result{}, err
}

func (s *containerResourceSyncer) SyncDown(ctx *synccontext.SyncContext, vObj client.Object) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// IsManaged determines if a physical object is managed by the vcluster
func (s *containerResourceSyncer) IsManaged(pObj client.Object) (bool, error) {
	return false, nil
}

type monotonicBool struct {
	modified bool
}

func (s *containerResourceSyncer) updateContainerResources(pObj, vObj *corev1.Pod) *corev1.Pod {
	updated := vObj.DeepCopy()
	if updated.Annotations == nil {
		updated.Annotations = map[string]string{}
	}

	b := &monotonicBool{}
	for i, c := range pObj.Spec.Containers {
		for _, resource := range []string{"limits", "requests"} {
			cpu := fmt.Sprintf("hostCluster.%s.%s.cpu", c.Name, resource)
			memory := fmt.Sprintf("hostCluster.%s.%s.memory", c.Name, resource)
			storage := fmt.Sprintf("hostCluster.%s.%s.storage", c.Name, resource)
			ephemeralStorage := fmt.Sprintf("hostCluster.%s.%s.ephemeral-storage", c.Name, resource)

			if resource == "limits" {
				noLimits := vObj.Spec.Containers[i].Resources.Limits == nil
				if noLimits || vObj.Spec.Containers[i].Resources.Limits.Cpu() == nil {
					updateMap(updated.Annotations, cpu, c.Resources.Limits.Cpu().String(), b)
				}
				if noLimits || vObj.Spec.Containers[i].Resources.Limits.Memory() == nil {
					updateMap(updated.Annotations, memory, c.Resources.Limits.Memory().String(), b)
				}
				if noLimits || vObj.Spec.Containers[i].Resources.Limits.Storage() == nil {
					updateMap(updated.Annotations, storage, c.Resources.Limits.Storage().String(), b)
				}
				if noLimits || vObj.Spec.Containers[i].Resources.Limits.StorageEphemeral() == nil {
					updateMap(updated.Annotations, ephemeralStorage, c.Resources.Limits.StorageEphemeral().String(), b)
				}
			} else { // requests
				noRequests := vObj.Spec.Containers[i].Resources.Requests == nil
				if noRequests || vObj.Spec.Containers[i].Resources.Requests.Cpu() == nil {
					updateMap(updated.Annotations, cpu, c.Resources.Requests.Cpu().String(), b)
				}
				if noRequests || vObj.Spec.Containers[i].Resources.Requests.Memory() == nil {
					updateMap(updated.Annotations, memory, c.Resources.Requests.Memory().String(), b)
				}
				if noRequests || vObj.Spec.Containers[i].Resources.Requests.Storage() == nil {
					updateMap(updated.Annotations, storage, c.Resources.Requests.Storage().String(), b)
				}
				if noRequests || vObj.Spec.Containers[i].Resources.Requests.StorageEphemeral() == nil {
					updateMap(updated.Annotations, ephemeralStorage, c.Resources.Requests.StorageEphemeral().String(), b)
				}
			}
		}
	}

	if !b.modified {
		return nil
	}
	return updated
}

func updateMap(strMap map[string]string, key, value string, b *monotonicBool) {
	_, found := strMap[key]
	b.modified = !found || b.modified
	strMap[key] = value
}
