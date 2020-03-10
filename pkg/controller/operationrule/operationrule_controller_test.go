package operationrule

import (
	"encoding/json"
	"strconv"
	"testing"

	oappsv1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/assert"
	rulev1alpha1 "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestFieldConversions ...
func TestFieldConversionsDep(t *testing.T) {
	priority := int32(5)
	termSec := int64(60)
	activeSec := int64(45)
	preempt := corev1.PreemptLowerPriority
	dep := appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deployment1",
			Labels: map[string]string{
				"test": "old",
				"ha":   "ha",
			},
		},
		Spec: appv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Priority:                      &priority,
					TerminationGracePeriodSeconds: &termSec,
					ActiveDeadlineSeconds:         &activeSec,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					ServiceAccountName:            "testsa",
					PreemptionPolicy:              &preempt,
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  "test",
									Value: "old",
								},
							},
						},
					},
				},
			},
		},
	}
	dep.SetGroupVersionKind(appv1.SchemeGroupVersion.WithKind("Deployment"))

	newPriority := priority - 2
	newTermSec := termSec * 2
	newActiveSec := activeSec * 2
	hostNetwork := true
	newSA := "newuser"
	v := rulev1alpha1.Variable{
		Name: dep.Name,
		Path: "spec.template.spec",
		Value: map[string]string{
			"activeDeadlineSeconds":         strconv.Itoa(int(newActiveSec)),
			"priority":                      strconv.Itoa(int(newPriority)),
			"terminationGracePeriodSeconds": strconv.Itoa(int(newTermSec)),
			"hostNetwork":                   strconv.FormatBool(hostNetwork),
			"restartPolicy":                 "OnFailure",
			"serviceAccountName":            newSA,
			"preemptionPolicy":              "Never",
		},
	}
	newJSON := fieldTypeConversion(v, dep.DeepCopyObject())

	newDep := &appv1.Deployment{}
	err := json.Unmarshal(newJSON, &newDep)
	assert.Nil(t, err)

	assert.NotEqual(t, dep.Spec.Template.Spec, newDep.Spec.Template.Spec)
	assert.Equal(t, &newPriority, newDep.Spec.Template.Spec.Priority)
	assert.Equal(t, &newTermSec, newDep.Spec.Template.Spec.TerminationGracePeriodSeconds)
	assert.Equal(t, &newActiveSec, newDep.Spec.Template.Spec.ActiveDeadlineSeconds)
	assert.Equal(t, hostNetwork, newDep.Spec.Template.Spec.HostNetwork)
	assert.Equal(t, corev1.RestartPolicyOnFailure, newDep.Spec.Template.Spec.RestartPolicy)
	assert.Equal(t, newSA, newDep.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(t, corev1.PreemptNever, *newDep.Spec.Template.Spec.PreemptionPolicy)

	v = rulev1alpha1.Variable{
		Name:  dep.Name,
		Path:  "metadata.labels",
		Value: map[string]string{"test": "new"},
	}

	newJSON = fieldTypeConversion(v, dep.DeepCopyObject())

	newDep = &appv1.Deployment{}
	err = json.Unmarshal(newJSON, &newDep)
	assert.Nil(t, err)
	assert.NotEqual(t, dep.ObjectMeta.Labels, newDep.ObjectMeta.Labels)
	assert.Equal(t, map[string]string{"test": "new", "ha": "ha"}, newDep.ObjectMeta.Labels)

	v = rulev1alpha1.Variable{
		Name:  dep.Name,
		Path:  "spec.template.spec.containers[0].env",
		Value: map[string]string{"testnew": "old", "test": "new"},
	}

	newJSON = fieldTypeConversion(v, dep.DeepCopyObject())

	newDep = &appv1.Deployment{}
	err = json.Unmarshal(newJSON, &newDep)
	assert.Nil(t, err)
	//assert.NotEqual(t, dep.Spec.Template.Spec.Containers[0].Env, newDep.Spec.Template.Spec.Containers[0].Env)
	//assert.Equal(t, []corev1.EnvVar{{Name: "testnew", Value: "old"}, {Name: "test", Value: "new"}}, newDep.Spec.Template.Spec.Containers[0].Env)
}

func TestFieldConversionsDC(t *testing.T) {
	priority := int32(5)
	termSec := int64(60)
	activeSec := int64(45)
	preempt := corev1.PreemptLowerPriority
	dc := oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deployment2",
			Labels: map[string]string{
				"test": "old",
				"ha":   "ha",
			},
		},
		Spec: oappsv1.DeploymentConfigSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Priority:                      &priority,
					TerminationGracePeriodSeconds: &termSec,
					ActiveDeadlineSeconds:         &activeSec,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					ServiceAccountName:            "testsa",
					PreemptionPolicy:              &preempt,
				},
			},
		},
	}
	dc.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))

	newPriority := priority - 2
	newTermSec := termSec * 2
	newActiveSec := activeSec * 2
	hostNetwork := true
	newSA := "newuser"
	v := rulev1alpha1.Variable{
		Name: dc.Name,
		Path: "spec.template.spec",
		Value: map[string]string{
			"activeDeadlineSeconds":         strconv.Itoa(int(newActiveSec)),
			"priority":                      strconv.Itoa(int(newPriority)),
			"terminationGracePeriodSeconds": strconv.Itoa(int(newTermSec)),
			"hostNetwork":                   strconv.FormatBool(hostNetwork),
			"restartPolicy":                 "OnFailure",
			"serviceAccountName":            newSA,
			"preemptionPolicy":              "Never",
			"config.yaml":                   "|\n    .defaults:\n      delete:\n        days: 10",
		},
	}
	newJSON := fieldTypeConversion(v, dc.DeepCopyObject())

	newDC := &appv1.Deployment{}
	err := json.Unmarshal(newJSON, &newDC)
	assert.Nil(t, err)

	assert.NotEqual(t, dc, newDC)
	assert.Equal(t, &newPriority, newDC.Spec.Template.Spec.Priority)
	assert.Equal(t, &newTermSec, newDC.Spec.Template.Spec.TerminationGracePeriodSeconds)
	assert.Equal(t, &newActiveSec, newDC.Spec.Template.Spec.ActiveDeadlineSeconds)
	assert.Equal(t, hostNetwork, newDC.Spec.Template.Spec.HostNetwork)
	assert.Equal(t, corev1.RestartPolicyOnFailure, newDC.Spec.Template.Spec.RestartPolicy)
	assert.Equal(t, newSA, newDC.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(t, corev1.PreemptNever, *newDC.Spec.Template.Spec.PreemptionPolicy)

	v = rulev1alpha1.Variable{
		Name:  dc.Name,
		Path:  "metadata.labels",
		Value: map[string]string{"test": "new"},
	}

	newJSON = fieldTypeConversion(v, dc.DeepCopyObject())

	newDC = &appv1.Deployment{}
	err = json.Unmarshal(newJSON, &newDC)
	assert.Nil(t, err)

	assert.NotEqual(t, dc, newDC)
	assert.Equal(t, map[string]string{"test": "new", "ha": "ha"}, newDC.ObjectMeta.Labels)
}

func TestFieldConversionsCM(t *testing.T) {
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "configMap",
		},
		Data: map[string]string{
			"config.yaml": "empty",
		},
	}
	cm.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	config := "|\n    .defaults:\n      delete:\n        days: 10"
	v := rulev1alpha1.Variable{
		Name:  cm.Name,
		Path:  "data",
		Value: map[string]string{`config\.yaml`: config},
	}
	newJSON := fieldTypeConversion(v, cm.DeepCopyObject())

	newCM := &corev1.ConfigMap{}
	err := json.Unmarshal(newJSON, &newCM)
	assert.Nil(t, err)

	convCMvalue := map[string]string{"config.yaml": config}
	assert.NotEqual(t, cm.Data, newCM.Data)
	assert.Equal(t, convCMvalue, newCM.Data)
}
