package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rsdenck/nux/internal/core/ports"
)

// ParseCrontabOutput parses the output of `crontab -l`
func ParseCrontabOutput(output string) ([]ports.CronJob, error) {
	var jobs []ports.CronJob
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		
		var schedule, command, user string
		
		if strings.HasPrefix(fields[0], "@") {
			if len(fields) < 2 {
				continue
			}
			schedule = fields[0]
			command = strings.Join(fields[1:], " ")
			user = detectUserFromCommand(command)
		} else {
			if len(fields) < 6 {
				continue
			}
			schedule = strings.Join(fields[:5], " ")
			command = strings.Join(fields[5:], " ")
			user = detectUserFromCommand(command)
		}

		jobs = append(jobs, ports.CronJob{
			ID:       fmt.Sprintf("cron-%d", i),
			Schedule: schedule,
			Command:  command,
			User:     user,
			File:     "user-crontab",
		})
	}
	return jobs, nil
}

func detectUserFromCommand(command string) string {
	if strings.Contains(command, "root") {
		return "root"
	}
	return "current"
}

// SystemdTimerEntry for JSON unmarshalling
type SystemdTimerEntry struct {
	Unit      string `json:"unit"`
	Activates string `json:"activates"`
	Next      string `json:"next"`
	Left      string `json:"left"`
	Last      string `json:"last"`
	Passed    string `json:"passed"`
}

// ParseSystemdTimersJSON parses the output of `systemctl list-timers --output=json`
func ParseSystemdTimersJSON(output string) ([]ports.SystemdTimer, error) {
	var entries []SystemdTimerEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		return nil, err
	}

	var timers []ports.SystemdTimer
	for _, entry := range entries {
		timers = append(timers, ports.SystemdTimer{
			Unit:    entry.Unit,
			Next:    entry.Next,
			Last:    entry.Last,
			Left:    entry.Left,
			Passed:  entry.Passed,
			Service: entry.Activates,
		})
	}
	return timers, nil
}
