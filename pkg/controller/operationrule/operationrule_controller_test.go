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
)

// TestFieldConversions ...
func TestFieldConversions(t *testing.T) {
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
	existingJSON, err := json.Marshal(&dep)
	assert.Nil(t, err)

	newPriority := priority - 2
	newTermSec := termSec * 2
	newActiveSec := activeSec * 2
	//hostNetwork := true
	v := rulev1alpha1.Variable{
		Name: dep.Name,
		Path: "spec.template.spec",
		Value: map[string]string{
			"activeDeadlineSeconds":         strconv.Itoa(int(newActiveSec)),
			"priority":                      strconv.Itoa(int(newPriority)),
			"terminationGracePeriodSeconds": strconv.Itoa(int(newTermSec)),
			//"hostNetwork":                   strconv.FormatBool(hostNetwork),
		},
	}
	newJSON := fieldTypeConversion(existingJSON, v)
	assert.NotEqual(t, existingJSON, newJSON)

	newdep := &appv1.Deployment{}
	err = json.Unmarshal(newJSON, newdep)
	assert.Nil(t, err)

	assert.Equal(t, newPriority, *newdep.Spec.Template.Spec.Priority)
	assert.Equal(t, newTermSec, *newdep.Spec.Template.Spec.TerminationGracePeriodSeconds)
	assert.Equal(t, newActiveSec, *newdep.Spec.Template.Spec.ActiveDeadlineSeconds)
	//assert.Equal(t, hostNetwork, newdep.Spec.Template.Spec.HostNetwork)
}
