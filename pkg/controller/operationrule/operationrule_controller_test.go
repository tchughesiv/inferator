package operationrule

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestFieldConversions ...
func TestFieldConversions(t *testing.T) {
	registerObjs := []runtime.Object{&rulev1alpha1.OperationRule{}, &rulev1alpha1.OperationRuleList{}, &appv1.Deployment{}}
	rulev1alpha1.SchemeBuilder.Register(registerObjs...)
	scheme, _ := rulev1alpha1.SchemeBuilder.Build()

	priority := int32(5)
	termSec := int64(60)
	activeSec := int64(45)
	dep := appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deployment2",
		},
		Spec: appv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Priority:                      &priority,
					TerminationGracePeriodSeconds: &termSec,
					ActiveDeadlineSeconds:         &activeSec,
				},
			},
		},
	}
	newPriority := priority - 2
	newTermSec := termSec * 2
	newActiveSec := activeSec * 2
	hostNetwork := true
	v := rulev1alpha1.Variable{
		Name: dep.Name,
		Path: "spec.template.spec",
		Value: map[string]string{
			"activeDeadlineSeconds":         strconv.Itoa(int(newActiveSec)),
			"priority":                      strconv.Itoa(int(newPriority)),
			"terminationGracePeriodSeconds": strconv.Itoa(int(newTermSec)),
			"hostNetwork":                   strconv.FormatBool(hostNetwork),
		},
	}
	objectOut, _ := fieldTypeConversion(&dep, v, scheme)

	newJSON, err := json.Marshal(&objectOut)
	assert.Nil(t, err)

	newDep := &appv1.Deployment{}
	err = json.Unmarshal(newJSON, newDep)
	assert.Nil(t, err)

	assert.NotEqual(t, &dep, newDep)
	assert.Equal(t, dep.Name, newDep.Name)

	assert.Equal(t, newPriority, *newDep.Spec.Template.Spec.Priority)
	assert.Equal(t, &newPriority, newDep.Spec.Template.Spec.Priority)

	assert.Equal(t, newTermSec, *newDep.Spec.Template.Spec.TerminationGracePeriodSeconds)
	assert.Equal(t, &newTermSec, newDep.Spec.Template.Spec.TerminationGracePeriodSeconds)

	assert.Equal(t, newActiveSec, *newDep.Spec.Template.Spec.ActiveDeadlineSeconds)
	assert.Equal(t, &newActiveSec, newDep.Spec.Template.Spec.ActiveDeadlineSeconds)

	//assert.Equal(t, hostNetwork, newDep.Spec.Template.Spec.HostNetwork)
}
