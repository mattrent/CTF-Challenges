package infrastructure

import (
	"context"
	"deployer/config"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func StartCleaner() {
	kubeconfig := GetKubeConfigSingleton()

	for range time.Tick(time.Minute * 1) {
		clientset, err := kubernetes.NewForConfig(kubeconfig)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Println(err.Error())
			continue
		}

		log.Println("Checking for challenges to remove")

		for _, ns := range nsList.Items {
			if strings.HasPrefix(ns.Name, challengeNamespacePrefix) {
				if ns.CreationTimestamp.Time.Add(time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)).Before(time.Now()) {
					log.Println("Deleting: " + ns.Name)
					err = clientset.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
					if err != nil {
						log.Println(err.Error())
						continue
					}
				}
			}
		}
	}
}
