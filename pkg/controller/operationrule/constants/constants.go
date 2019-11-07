package constants

import corev1 "k8s.io/api/core/v1"

// DebugTrue - used to enable debug logs in objects
var DebugTrue = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "true",
}

// DebugFalse - used to disable debug logs in objects
var DebugFalse = corev1.EnvVar{
	Name:  "DEBUG",
	Value: "false",
}
