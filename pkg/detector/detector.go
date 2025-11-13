package detector

import (
	"context"
	"time"

	"github.com/rrdesai64/zombie-hunter/pkg/k8s"
)

type Detector struct {
	client *k8s.Client
}

func NewDetector(client *k8s.Client) *Detector {
	return &Detector{client: client}
}

type ZombieCandidate struct {
	Name             string
	Namespace        string
	Schedule         string
	LastSuccess      *time.Time
	DaysSinceSuccess int
	TotalJobs        int
	FailedJobs       int
	Confidence       int
	IsSuspended      bool
}

// FindZombies detects zombie CronJobs
func (d *Detector) FindZombies(ctx context.Context, namespace string, thresholdDays int) ([]ZombieCandidate, error) {
	cronJobs, err := d.client.GetCronJobs(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var zombies []ZombieCandidate

	for _, cj := range cronJobs {
		candidate := d.analyzeCronJob(ctx, cj)

		if candidate.DaysSinceSuccess >= thresholdDays || candidate.TotalJobs == 0 {
			zombies = append(zombies, candidate)
		}
	}

	return zombies, nil
}

func (d *Detector) analyzeCronJob(ctx context.Context, cronJob k8s.CronJobInfo) ZombieCandidate {
	jobs, err := d.client.GetJobsForCronJob(ctx, cronJob.Namespace, cronJob.Name)
	if err != nil {
		return ZombieCandidate{
			Name:             cronJob.Name,
			Namespace:        cronJob.Namespace,
			Schedule:         cronJob.Schedule,
			DaysSinceSuccess: 999,
			TotalJobs:        0,
			FailedJobs:       0,
			Confidence:       50,
			IsSuspended:      cronJob.Suspended,
		}
	}

	var lastSuccess *time.Time
	failedCount := 0
	totalCount := len(jobs)

	for _, job := range jobs {
		if job.Succeeded > 0 && job.CompletionTime != nil {
			if lastSuccess == nil || job.CompletionTime.After(*lastSuccess) {
				lastSuccess = job.CompletionTime
			}
		}
		if job.Failed > 0 {
			failedCount++
		}
	}

	daysSince := 999
	if lastSuccess != nil {
		daysSince = int(time.Since(*lastSuccess).Hours() / 24)
	}

	confidence := calculateConfidence(daysSince, totalCount, failedCount, cronJob.Suspended)

	return ZombieCandidate{
		Name:             cronJob.Name,
		Namespace:        cronJob.Namespace,
		Schedule:         cronJob.Schedule,
		LastSuccess:      lastSuccess,
		DaysSinceSuccess: daysSince,
		TotalJobs:        totalCount,
		FailedJobs:       failedCount,
		Confidence:       confidence,
		IsSuspended:      cronJob.Suspended,
	}
}

func calculateConfidence(daysSince, total, failed int, suspended bool) int {
	if suspended {
		return 20
	}

	if total == 0 {
		return 50
	}

	if total > 0 && failed == total {
		return 95
	}

	if daysSince >= 365 {
		return 99
	}
	if daysSince >= 180 {
		return 95
	}
	if daysSince >= 90 {
		return 85
	}
	if daysSince >= 60 {
		return 75
	}
	if daysSince >= 30 {
		return 60
	}

	return 40
}
