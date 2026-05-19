// Package scheduler provides cron job and systemd timer parsing capabilities.
//
// This package includes:
//   - ParseCrontabOutput: Parse crontab format strings into CronJob structs
//   - ParseSystemdTimersJSON: Parse systemd timer JSON output into SystemdTimer structs
//
// Example usage:
//
//	jobs, err := scheduler.ParseCrontabOutput("0 5 * * * /usr/bin/backup.sh")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, job := range jobs {
//	    fmt.Printf("Schedule: %s, Command: %s\n", job.Schedule, job.Command)
//	}
package scheduler