package constants

import corev1 "k8s.io/api/core/v1"

const (
	RuntimeEnv = "INFERATOR"
	RulesVar   = "RULES_DEFINITION"
)

var (
	DebugTrue = corev1.EnvVar{
		Name:  "DEBUG",
		Value: "true",
	}
	DebugFalse = corev1.EnvVar{
		Name:  "DEBUG",
		Value: "false",
	}
)
