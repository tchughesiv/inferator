package operationrule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/alertmanager/template"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (reconciler *Reconciler) startWebhook() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/alertwebhook", reconciler.alertWebhook)
	listenAddress := ":8081"
	log.Info("webhook listening, ", "addr: ", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func (reconciler *Reconciler) alertWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		asJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	// fmt.Printf("Alerts: GroupLabels=%v, CommonLabels=%v\n", data.GroupLabels, data.CommonLabels)
	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: os.Getenv("WATCH_NAMESPACE")}}
	for _, alert := range data.Alerts {
		reconciler.reconcileInferator(request, alert)
	}
	asJSON(w, http.StatusOK, "success")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Fprint(w, "Ok!")
}

type responseJSON struct {
	Status  int
	Message string
}

func asJSON(w http.ResponseWriter, status int, message string) {
	data := responseJSON{
		Status:  status,
		Message: message,
	}
	bytes, _ := json.Marshal(data)
	json := string(bytes[:])

	w.WriteHeader(status)
	fmt.Fprint(w, json)
}

/*
	func sms(alert template.Alert) {
		fmt.Printf("sending sms through twilio...")
		twilio := gotwilio.NewTwilioClient(os.Getenv("TWILIO_ACCOUNT"), os.Getenv("TWILIO_TOKEN"))
		from := os.Getenv("TWILIO_FROM")
		message := alert.Annotations["summary"] + " Status: " + alert.Status
		to := alert.Annotations["phones"]
		reg := regexp.MustCompile("\\s*,\\s*")
		for _, r := range reg.Split(to, -1) {
			if strings.TrimSpace(r) != "" {
				_, _, err := twilio.SendSMS(from, r, message, "", "")
				if err != nil {
					fmt.Printf("error sending SMS: %v", err)
				}
			} else {
				fmt.Printf("ignore empty recipient")
			}
		}
	}
*/
