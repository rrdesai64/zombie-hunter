package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rrdesai64/zombie-hunter/pkg/detector"
)

type Formatter struct {
	format string
}

func NewFormatter(format string) *Formatter {
	return &Formatter{format: format}
}

func (f *Formatter) Output(zombies []detector.ZombieCandidate, thresholdDays int) error {
	switch f.format {
	case "json":
		return f.outputJSON(zombies)
	case "csv":
		return f.outputCSV(zombies)
	default:
		return f.outputTable(zombies, thresholdDays)
	}
}

func (f *Formatter) outputTable(zombies []detector.ZombieCandidate, thresholdDays int) error {
	fmt.Printf("\n🧟 ZOMBIE HUNTER REPORT\n")
	fmt.Printf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Threshold: %d days\n\n", thresholdDays)

	if len(zombies) == 0 {
		fmt.Printf("✅ No zombies found! All CronJobs are healthy.\n\n")
		return nil
	}

	fmt.Printf("%s\n", strings.Repeat("━", 80))
	fmt.Printf("ZOMBIE CANDIDATES (%d found)\n", len(zombies))
	fmt.Printf("%s\n\n", strings.Repeat("━", 80))

	// Simple table output
	fmt.Printf("%-4s %-30s %-15s %-15s %-12s %-20s\n",
		"🔍", "NAME", "NAMESPACE", "DAYS INACTIVE", "CONFIDENCE", "JOBS")
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	highConf := 0
	for _, z := range zombies {
		emoji := getEmoji(z.Confidence)

		daysStr := fmt.Sprintf("%d", z.DaysSinceSuccess)
		if z.DaysSinceSuccess >= 999 {
			daysStr = "NEVER"
		}

		jobsStr := fmt.Sprintf("%d total, %d failed", z.TotalJobs, z.FailedJobs)
		if z.IsSuspended {
			jobsStr += " (susp.)"
		}

		// Truncate long names
		name := z.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		fmt.Printf("%-4s %-30s %-15s %-15s %-12s %-20s\n",
			emoji,
			name,
			z.Namespace,
			daysStr,
			fmt.Sprintf("%d%%", z.Confidence),
			jobsStr,
		)

		if z.Confidence >= 80 {
			highConf++
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("━", 80))
	fmt.Printf("SUMMARY\n")
	fmt.Printf("%s\n\n", strings.Repeat("━", 80))

	fmt.Printf("Total zombies found: %d\n", len(zombies))
	fmt.Printf("High confidence (≥80%%): %d\n", highConf)

	if highConf > 0 {
		fmt.Printf("\n💡 Tip: Start by reviewing high-confidence zombies\n")
	}

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Review each zombie with your team\n")
	fmt.Printf("2. Delete safely: kubectl delete cronjob <name> -n <namespace>\n")
	fmt.Printf("3. Try different thresholds: --days 60 or --days 90\n\n")

	return nil
}

func (f *Formatter) outputCSV(zombies []detector.ZombieCandidate) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	w.Write([]string{"Name", "Namespace", "Schedule", "DaysSinceSuccess", "TotalJobs", "FailedJobs", "Confidence", "Suspended"})

	for _, z := range zombies {
		w.Write([]string{
			z.Name,
			z.Namespace,
			z.Schedule,
			fmt.Sprintf("%d", z.DaysSinceSuccess),
			fmt.Sprintf("%d", z.TotalJobs),
			fmt.Sprintf("%d", z.FailedJobs),
			fmt.Sprintf("%d", z.Confidence),
			fmt.Sprintf("%v", z.IsSuspended),
		})
	}

	return nil
}

func (f *Formatter) outputJSON(zombies []detector.ZombieCandidate) error {
	output := map[string]interface{}{
		"generated_at":  time.Now().Format(time.RFC3339),
		"total_zombies": len(zombies),
		"zombies":       zombies,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func getEmoji(confidence int) string {
	if confidence >= 90 {
		return "💀"
	} else if confidence >= 70 {
		return "⚠️"
	} else if confidence >= 50 {
		return "🤔"
	} else {
		return "ℹ️"
	}
}
