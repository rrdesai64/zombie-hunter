package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rrdesai64/zombie-hunter/pkg/detector"
	"github.com/rrdesai64/zombie-hunter/pkg/k8s"
	"github.com/rrdesai64/zombie-hunter/pkg/report"
	"github.com/spf13/cobra"
)

var (
	days      int
	namespace string
	format    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "zombie-hunter",
		Short: "Find zombie CronJobs in your Kubernetes cluster",
		Long: `Zombie Hunter scans your Kubernetes cluster for CronJobs that haven't 
run successfully in a specified number of days. These "zombies" cost you money
and clutter your infrastructure.`,
		RunE: run,
	}

	rootCmd.Flags().IntVar(&days, "days", 30, "Consider zombie if no success in N days")
	rootCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace (empty = all)")
	rootCmd.Flags().StringVar(&format, "format", "table", "Output format: table, csv, json")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create K8s client
	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get CronJobs
	cronJobsList, err := client.GetRawCronJobs(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to list CronJobs: %w", err)
	}

	// Find zombies
	var zombies []detector.Zombie

	for _, cronJob := range cronJobsList.Items {
		// Get Jobs for this CronJob
		jobsList, err := client.GetRawJobsForCronJob(ctx, cronJob.Namespace, cronJob.Name)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get jobs for %s/%s: %v\n",
				cronJob.Namespace, cronJob.Name, err)
			continue
		}

		// Analyze this CronJob
		zombie := detector.AnalyzeCronJob(&cronJob, jobsList.Items, days)

		if zombie.IsZombie {
			zombies = append(zombies, zombie)
		}
	}

	// Format and output
	formatter := report.NewFormatter(format)
	return formatter.Output(zombies, days)
}
