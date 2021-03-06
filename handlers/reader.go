package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getServiceList(clientset *kubernetes.Clientset) ([]requests.Function, error) {
	var functions []requests.Function

	listOpts := metav1.ListOptions{
		LabelSelector: "faas_function",
	}

	res, err := clientset.ExtensionsV1beta1().Deployments("default").List(listOpts)

	if err != nil {
		return nil, err
	}
	for _, item := range res.Items {
		var replicas uint64
		if item.Spec.Replicas != nil {
			replicas = uint64(*item.Spec.Replicas)
		}

		function := requests.Function{
			Name:            item.Name,
			Replicas:        replicas,
			Image:           item.Spec.Template.Spec.Containers[0].Image,
			InvocationCount: 0,
		}
		functions = append(functions, function)
	}
	return functions, nil
}

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		functions, err := getServiceList(clientset)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}
