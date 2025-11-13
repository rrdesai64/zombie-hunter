package k8s

import (
	"context"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{clientset: clientset}, nil
}

// getConfig returns K8s config
func getConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// GetCronJobs returns all CronJobs
func (c *Client) GetCronJobs(ctx context.Context, namespace string) ([]CronJobInfo, error) {
	var ns string
	if namespace == "" {
		ns = metav1.NamespaceAll
	} else {
		ns = namespace
	}

	cronJobs, err := c.clientset.BatchV1().CronJobs(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []CronJobInfo
	for _, cj := range cronJobs.Items {
		info := CronJobInfo{
			Name:      cj.Name,
			Namespace: cj.Namespace,
			Schedule:  cj.Spec.Schedule,
			Suspended: false,
		}

		if cj.Spec.Suspend != nil {
			info.Suspended = *cj.Spec.Suspend
		}

		result = append(result, info)
	}

	return result, nil
}

// GetJobsForCronJob returns all Jobs for a CronJob
func (c *Client) GetJobsForCronJob(ctx context.Context, namespace, cronJobName string) ([]JobInfo, error) {
	jobs, err := c.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "cronjob=" + cronJobName,
	})
	if err != nil {
		return nil, err
	}

	var result []JobInfo
	for _, job := range jobs.Items {
		info := JobInfo{
			Name:           job.Name,
			Succeeded:      job.Status.Succeeded,
			Failed:         job.Status.Failed,
			CompletionTime: nil,
		}

		if job.Status.CompletionTime != nil {
			t := job.Status.CompletionTime.Time
			info.CompletionTime = &t
		}

		result = append(result, info)
	}

	return result, nil
}

type CronJobInfo struct {
	Name      string
	Namespace string
	Schedule  string
	Suspended bool
}

type JobInfo struct {
	Name           string
	Succeeded      int32
	Failed         int32
	CompletionTime *time.Time
}
