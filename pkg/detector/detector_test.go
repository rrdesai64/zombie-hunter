package detector

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name       string
		daysSince  int
		totalJobs  int
		failedJobs int
		suspended  bool
		expected   int
	}{
		{
			name:       "Suspended job should have low confidence",
			daysSince:  100,
			totalJobs:  10,
			failedJobs: 0,
			suspended:  true,
			expected:   20,
		},
		{
			name:       "Never ran should be 50% confidence",
			daysSince:  100,
			totalJobs:  0,
			failedJobs: 0,
			suspended:  false,
			expected:   50,
		},
		{
			name:       "All jobs failed should be 95% confidence",
			daysSince:  50,
			totalJobs:  10,
			failedJobs: 10,
			suspended:  false,
			expected:   95,
		},
		{
			name:       "365+ days inactive should be 99%",
			daysSince:  400,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   99,
		},
		{
			name:       "180-365 days should be 95%",
			daysSince:  200,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   95,
		},
		{
			name:       "90-180 days should be 85%",
			daysSince:  120,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   85,
		},
		{
			name:       "60-90 days should be 75%",
			daysSince:  75,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   75,
		},
		{
			name:       "30-60 days should be 60%",
			daysSince:  45,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   60,
		},
		{
			name:       "Less than 30 days should be 40%",
			daysSince:  20,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   40,
		},
		{
			name:       "Edge case: exactly 365 days",
			daysSince:  365,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   99,
		},
		{
			name:       "Edge case: exactly 30 days",
			daysSince:  30,
			totalJobs:  5,
			failedJobs: 0,
			suspended:  false,
			expected:   60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateConfidence(tt.daysSince, tt.totalJobs, tt.failedJobs, tt.suspended)
			if result != tt.expected {
				t.Errorf("CalculateConfidence(%d, %d, %d, %v) = %d; want %d",
					tt.daysSince, tt.totalJobs, tt.failedJobs, tt.suspended, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeCronJob(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name               string
		cronJob            *batchv1.CronJob
		jobs               []batchv1.Job
		thresholdDays      int
		expectedIsZombie   bool
		expectedConfidence int
	}{
		{
			name: "CronJob with recent successful job",
			cronJob: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "active-job",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "0 0 * * *",
				},
			},
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-24 * time.Hour)),
							},
						},
					},
				},
			},
			thresholdDays:      30,
			expectedIsZombie:   false,
			expectedConfidence: 0,
		},
		{
			name: "CronJob with old successful job - is zombie",
			cronJob: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "old-job",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "0 0 * * *",
				},
			},
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-100 * 24 * time.Hour)),
							},
						},
					},
				},
			},
			thresholdDays:      30,
			expectedIsZombie:   true,
			expectedConfidence: 85, // 90-180 days
		},
		{
			name: "Suspended CronJob should have low confidence",
			cronJob: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "suspended-job",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "0 0 * * *",
					Suspend:  boolPtr(true),
				},
			},
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-100 * 24 * time.Hour)),
							},
						},
					},
				},
			},
			thresholdDays:      30,
			expectedIsZombie:   true,
			expectedConfidence: 20, // Suspended
		},
		{
			name: "CronJob with no jobs should be zombie with 50% confidence",
			cronJob: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-jobs",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "0 0 * * *",
				},
			},
			jobs:               []batchv1.Job{},
			thresholdDays:      30,
			expectedIsZombie:   true,
			expectedConfidence: 50,
		},
		{
			name: "CronJob with all failed jobs should be high confidence zombie",
			cronJob: &batchv1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "failed-jobs",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "0 0 * * *",
				},
			},
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobFailed,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-50 * 24 * time.Hour)),
							},
						},
					},
				},
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobFailed,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-40 * 24 * time.Hour)),
							},
						},
					},
				},
			},
			thresholdDays:      30,
			expectedIsZombie:   true,
			expectedConfidence: 95, // All failed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zombie := AnalyzeCronJob(tt.cronJob, tt.jobs, tt.thresholdDays)

			if zombie.IsZombie != tt.expectedIsZombie {
				t.Errorf("IsZombie = %v; want %v", zombie.IsZombie, tt.expectedIsZombie)
			}

			if zombie.IsZombie && zombie.Confidence != tt.expectedConfidence {
				t.Errorf("Confidence = %d; want %d", zombie.Confidence, tt.expectedConfidence)
			}
		})
	}
}

func TestDaysSinceSuccess(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		jobs     []batchv1.Job
		expected int
	}{
		{
			name: "Recent successful job",
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-5 * 24 * time.Hour)),
							},
						},
					},
				},
			},
			expected: 5,
		},
		{
			name: "No successful jobs",
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobFailed,
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: 999,
		},
		{
			name:     "No jobs at all",
			jobs:     []batchv1.Job{},
			expected: 999,
		},
		{
			name: "Multiple jobs, only care about most recent success",
			jobs: []batchv1.Job{
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-100 * 24 * time.Hour)),
							},
						},
					},
				},
				{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(now.Add(-10 * 24 * time.Hour)),
							},
						},
					},
				},
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DaysSinceSuccess(tt.jobs) // ✅ UPPERCASE!
			if result != tt.expected {
				t.Errorf("DaysSinceSuccess() = %d; want %d", result, tt.expected)
			}
		})
	}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
