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
			shouldDelete := false
			if strings.HasPrefix(ns.Name, challengeNamespacePrefix) {

				// Check if the namespace has expired
				if ns.CreationTimestamp.Time.Add(time.Minute * time.Duration(config.Values.ChallengeLifetimeMinutes)).Before(time.Now()) {
					shouldDelete = true
				} else {
					// Check if all resources in the namespace are completed or no longer running
					podList, err := clientset.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{})
					if err != nil {
						log.Println(err.Error())
						continue
					}

					// TODO verify that this works
					allCompleted := true
					for _, pod := range podList.Items {
						if pod.Status.Phase != "Succeeded" && pod.Status.Phase != "Failed" {
							allCompleted = false
							break
						}
					}

					if allCompleted {
						shouldDelete = true
					}
				}
			}

			if shouldDelete {
				log.Println("Deleting namespace: " + ns.Name)
				err = clientset.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Println(err.Error())
					continue
				}
			}
		}
	}
}
