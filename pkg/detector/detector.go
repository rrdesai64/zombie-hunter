package detector

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// Zombie represents a potentially abandoned CronJob
type Zombie struct {
	Name             string
	Namespace        string
	Schedule         string
	DaysSinceSuccess int
	Confidence       int
	TotalJobs        int
	FailedJobs       int
	IsSuspended      bool
	IsZombie         bool
}

// AnalyzeCronJob analyzes a CronJob and its Jobs to determine if it's a zombie
func AnalyzeCronJob(cronJob *batchv1.CronJob, jobs []batchv1.Job, thresholdDays int) Zombie {
	daysSince := DaysSinceSuccess(jobs)
	totalJobs := len(jobs)
	failedJobs := countFailedJobs(jobs)
	isSuspended := cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend

	zombie := Zombie{
		Name:             cronJob.Name,
		Namespace:        cronJob.Namespace,
		Schedule:         cronJob.Spec.Schedule,
		DaysSinceSuccess: daysSince,
		TotalJobs:        totalJobs,
		FailedJobs:       failedJobs,
		IsSuspended:      isSuspended,
		IsZombie:         false,
		Confidence:       0,
	}

	// Determine if it's a zombie
	if daysSince >= thresholdDays || totalJobs == 0 {
		zombie.IsZombie = true
		zombie.Confidence = CalculateConfidence(daysSince, totalJobs, failedJobs, isSuspended)
	}

	return zombie
}

// CalculateConfidence calculates confidence score (0-99%) that a CronJob is abandoned
func CalculateConfidence(daysSince, totalJobs, failedJobs int, suspended bool) int {
	// Suspended jobs are intentionally paused - low confidence
	if suspended {
		return 20
	}

	// Never ran - could be new or abandoned
	if totalJobs == 0 {
		return 50
	}

	// All jobs failed - clearly broken
	if totalJobs > 0 && failedJobs == totalJobs {
		return 95
	}

	// Time-based confidence scoring
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

// DaysSinceSuccess calculates days since last successful job completion
func DaysSinceSuccess(jobs []batchv1.Job) int {
	if len(jobs) == 0 {
		return 999 // Never ran
	}

	var lastSuccess *time.Time

	for _, job := range jobs {
		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
				if lastSuccess == nil || condition.LastTransitionTime.Time.After(*lastSuccess) {
					t := condition.LastTransitionTime.Time
					lastSuccess = &t
				}
			}
		}
	}

	if lastSuccess == nil {
		return 999 // No successful jobs
	}

	days := int(time.Since(*lastSuccess).Hours() / 24)
	return days
}

// countFailedJobs counts how many jobs have failed
func countFailedJobs(jobs []batchv1.Job) int {
	count := 0
	for _, job := range jobs {
		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobFailed && condition.Status == v1.ConditionTrue {
				count++
				break
			}
		}
	}
	return count
}
