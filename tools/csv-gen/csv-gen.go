package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/blang/semver"
	routev1 "github.com/openshift/api/route/v1"
	security1 "github.com/openshift/api/security/v1"
	csvv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olmversion "github.com/operator-framework/operator-lifecycle-manager/pkg/lib/version"
	api "github.com/tchughesiv/inferator/pkg/apis/rule/v1alpha1"
	"github.com/tchughesiv/inferator/pkg/components"
	"github.com/tchughesiv/inferator/tools/util"
	"github.com/tchughesiv/inferator/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	rh       = "Red Hat, Inc."
	maturity = "alpha"
)

type csvPermissions struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Rules              []rbacv1.PolicyRule `json:"rules"`
}
type csvDeployments struct {
	Name string                `json:"name"`
	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
}
type csvStrategySpec struct {
	Permissions []csvPermissions `json:"permissions"`
	Deployments []csvDeployments `json:"deployments"`
}
type channel struct {
	Name       string `json:"name"`
	CurrentCSV string `json:"currentCSV"`
}
type packageStruct struct {
	PackageName string    `json:"packageName"`
	Channels    []channel `json:"channels"`
}

func main() {
	csv := components.Csv
	operatorName := csv.Name + "-operator"
	templateStruct := &csvv1alpha1.ClusterServiceVersion{}
	templateStruct.SetGroupVersionKind(csvv1alpha1.SchemeGroupVersion.WithKind("ClusterServiceVersion"))
	csvStruct := &csvv1alpha1.ClusterServiceVersion{}
	strategySpec := &csvStrategySpec{}
	json.Unmarshal(csvStruct.Spec.InstallStrategy.StrategySpecRaw, strategySpec)

	templateStrategySpec := &csvStrategySpec{}
	deployment := components.GetDeployment(csv.OperatorName, csv.Registry, csv.Context, csv.ImageName, csv.Tag, "Always")
	templateStrategySpec.Deployments = append(templateStrategySpec.Deployments, []csvDeployments{{Name: csv.OperatorName, Spec: deployment.Spec}}...)
	role := components.GetRole(csv.OperatorName)
	templateStrategySpec.Permissions = append(templateStrategySpec.Permissions, []csvPermissions{{ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName, Rules: role.Rules}}...)
	// Re-serialize deployments and permissions into csv strategy.
	updatedStrat, err := json.Marshal(templateStrategySpec)
	if err != nil {
		panic(err)
	}
	templateStruct.Spec.InstallStrategy.StrategySpecRaw = updatedStrat
	templateStruct.Spec.InstallStrategy.StrategyName = "deployment"
	csvVersionedName := operatorName + "." + version.Version
	templateStruct.Name = csvVersionedName
	templateStruct.Namespace = "placeholder"
	descrip := csv.DisplayName + " " + version.Version
	repository := "https://github.com/tchughesiv/inferator"
	//examples := []string{"{\x22apiVersion\x22:\x22app.kiegroup.org/v2\x22,\x22kind\x22:\x22KieApp\x22,\x22metadata\x22:{\x22name\x22:\x22rhpam-trial\x22},\x22spec\x22:{\x22environment\x22:\x22rhpam-trial\x22}}"}
	templateStruct.SetAnnotations(
		map[string]string{
			"createdAt":           time.Now().Format("2006-01-02 15:04:05"),
			"containerImage":      deployment.Spec.Template.Spec.Containers[0].Image,
			"description":         descrip,
			"categories":          "Integration & Delivery",
			"certified":           "false",
			"capabilities":        "Basic Install",
			"repository":          repository,
			"support":             rh,
			"tectonic-visibility": "ocs",
			//"alm-examples":        "[" + strings.Join(examples, ",") + "]",
		},
	)
	templateStruct.SetLabels(
		map[string]string{
			"operator-" + csv.Name: "true",
		},
	)
	templateStruct.Spec.Keywords = []string{"inferator", "operator"}
	var opVersion olmversion.OperatorVersion
	opVersion.Version = semver.MustParse(version.Version)
	templateStruct.Spec.Version = opVersion
	templateStruct.Spec.Description = descrip
	templateStruct.Spec.DisplayName = csv.DisplayName
	templateStruct.Spec.Maturity = maturity
	templateStruct.Spec.Maintainers = []csvv1alpha1.Maintainer{{Name: rh, Email: "tohughes@redhat.com"}}
	templateStruct.Spec.Provider = csvv1alpha1.AppLink{Name: rh}
	tLabels := map[string]string{
		"alm-owner-" + csv.Name: operatorName,
		"operated-by":           csvVersionedName,
	}
	templateStruct.Spec.Labels = tLabels
	templateStruct.Spec.Selector = &metav1.LabelSelector{MatchLabels: tLabels}
	templateStruct.Spec.InstallModes = []csvv1alpha1.InstallMode{
		{Type: csvv1alpha1.InstallModeTypeOwnNamespace, Supported: true},
		{Type: csvv1alpha1.InstallModeTypeSingleNamespace, Supported: true},
		{Type: csvv1alpha1.InstallModeTypeMultiNamespace, Supported: false},
		{Type: csvv1alpha1.InstallModeTypeAllNamespaces, Supported: true},
	}
	templateStruct.Spec.CustomResourceDefinitions.Owned = []csvv1alpha1.CRDDescription{
		{
			Version:     api.SchemeGroupVersion.Version,
			Kind:        "OperationRule",
			DisplayName: "OperationRule",
			Description: "A project prescription running an Inferator pod.",
			Name:        "operationrules." + api.SchemeGroupVersion.Group,
			Resources: []csvv1alpha1.APIResourceReference{
				{
					Kind:    "Role",
					Version: rbacv1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "RoleBinding",
					Version: rbacv1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "ClusterRole",
					Version: rbacv1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "ClusterRoleBinding",
					Version: rbacv1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "Secret",
					Version: corev1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "Pod",
					Version: corev1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "ServiceAccount",
					Version: corev1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "Service",
					Version: corev1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "Route",
					Version: routev1.SchemeGroupVersion.String(),
				},
				{
					Kind:    "SecurityContextConstraint",
					Version: security1.SchemeGroupVersion.String(),
				},
			},
		},
	}

	templateStruct.Annotations["certified"] = "false"
	deployFile := "deploy/operator.yaml"
	createFile(deployFile, deployment)
	roleFile := "deploy/role.yaml"
	createFile(roleFile, role)
	csvFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + version.Version + "/" + csvVersionedName + ".clusterserviceversion.yaml"
	/*
		copyTemplateStruct := templateStruct.DeepCopy()
		copyTemplateStruct.Annotations["createdAt"] = ""
		data := &csvv1alpha1.ClusterServiceVersion{}
		if fileExists(csvFile) {
			yamlFile, err := ioutil.ReadFile(csvFile)
			if err != nil {
				log.Printf("yamlFile.Get err   #%v ", err)
			}
			err = yaml.Unmarshal(yamlFile, data)
			if err != nil {
				log.Fatalf("Unmarshal: %v", err)
			}
			data.Annotations["createdAt"] = ""
		}
		if !reflect.DeepEqual(copyTemplateStruct.Spec, data.Spec) ||
			!reflect.DeepEqual(copyTemplateStruct.Annotations, data.Annotations) ||
			!reflect.DeepEqual(copyTemplateStruct.Labels, data.Labels) {

			createFile(csvFile, templateStruct)
		}
	*/
	createFile(csvFile, templateStruct)

	packageFile := "deploy/catalog_resources/" + csv.CsvDir + "/" + csv.Name + ".package.yaml"
	p, err := os.Create(packageFile)
	defer p.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	pwr := bufio.NewWriter(p)
	pwr.WriteString("#! package-manifest: " + csvFile + "\n")
	packagedata := packageStruct{
		PackageName: operatorName,
		Channels: []channel{
			{
				Name:       maturity,
				CurrentCSV: csvVersionedName,
			},
		},
	}
	util.MarshallObject(packagedata, pwr)
	pwr.Flush()
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createFile(filepath string, obj interface{}) {
	f, err := os.Create(filepath)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	writer := bufio.NewWriter(f)
	util.MarshallObject(obj, writer)
	writer.Flush()
}
