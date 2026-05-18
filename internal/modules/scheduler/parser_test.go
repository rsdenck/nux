package scheduler

import (
	"testing"
)

func TestParseCrontabOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantJobs int
		wantErr  bool
	}{
		{
			name: "valid crontab with multiple entries",
			input: `# Sample crontab
0 5 * * * /usr/bin/backup.sh
*/15 * * * * /usr/bin/monitor.sh
@reboot /usr/bin/startup.sh
30 2 * * 0 /usr/bin/weekly-cleanup.sh`,
			wantJobs: 4,
			wantErr:  false,
		},
		{
			name:     "empty crontab",
			input:    "",
			wantJobs: 0,
			wantErr:  false,
		},
		{
			name:     "only comments",
			input:    "# This is a comment\n# Another comment",
			wantJobs: 0,
			wantErr:  false,
		},
		{
			name: "mixed valid and invalid lines",
			input: `# Valid entry
0 * * * * /usr/bin/hourly.sh
# Invalid - missing fields
invalid line
@daily /usr/bin/daily.sh`,
			wantJobs: 2,
			wantErr:  false,
		},
		{
			name: "complex schedule",
			input: `0 0 1 * * /usr/bin/monthly.sh
0 */2 * * * /usr/bin/every2hours.sh
30 4 * * 1-5 /usr/bin/weekday.sh`,
			wantJobs: 3,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCrontabOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCrontabOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantJobs {
				t.Errorf("ParseCrontabOutput() got %d jobs, want %d", len(got), tt.wantJobs)
			}
		})
	}
}

func TestParseCrontabOutput_SpecificCases(t *testing.T) {
	input := `0 5 * * * /usr/bin/backup.sh`
	jobs, err := ParseCrontabOutput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	job := jobs[0]
	if job.Schedule != "0 5 * * *" {
		t.Errorf("Expected schedule '0 5 * * *', got '%s'", job.Schedule)
	}
	if job.Command != "/usr/bin/backup.sh" {
		t.Errorf("Expected command '/usr/bin/backup.sh', got '%s'", job.Command)
	}
	if job.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestParseCrontabOutput_AtCommands(t *testing.T) {
	input := `@reboot /usr/bin/startup.sh
@daily /usr/bin/daily.sh
@weekly /usr/bin/weekly.sh
@monthly /usr/bin/monthly.sh`

	jobs, err := ParseCrontabOutput(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(jobs) != 4 {
		t.Fatalf("Expected 4 jobs, got %d", len(jobs))
	}

	expectedSchedules := []string{"@reboot", "@daily", "@weekly", "@monthly"}
	for i, job := range jobs {
		if job.Schedule != expectedSchedules[i] {
			t.Errorf("Expected schedule '%s', got '%s'", expectedSchedules[i], job.Schedule)
		}
	}
}

func TestParseSystemdTimersJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTimers int
		wantErr  bool
	}{
		{
			name: "valid systemd timers JSON",
			input: `[
				{
					"unit": "daily-backup.timer",
					"activates": "daily-backup.service",
					"next": "Mon 2024-01-15 03:00:00 UTC",
					"left": "17h 23min",
					"last": "Sun 2024-01-14 03:00:01 UTC",
					"passed": "1 day ago"
				},
				{
					"unit": "hourly-cleanup.timer",
					"activates": "hourly-cleanup.service",
					"next": "Mon 2024-01-15 10:00:00 UTC",
					"left": "1h 23min",
					"last": "Mon 2024-01-15 09:00:01 UTC",
					"passed": "59min ago"
				}
			]`,
			wantTimers: 2,
			wantErr:  false,
		},
		{
			name:     "empty timers array",
			input:    "[]",
			wantTimers: 0,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    "{invalid json}",
			wantTimers: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSystemdTimersJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSystemdTimersJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantTimers {
				t.Errorf("ParseSystemdTimersJSON() got %d timers, want %d", len(got), tt.wantTimers)
			}
		})
	}
}

func TestParseSystemdTimersJSON_Fields(t *testing.T) {
	input := `[
		{
			"unit": "test.timer",
			"activates": "test.service",
			"next": "2024-01-15 03:00:00",
			"left": "10h",
			"last": "2024-01-14 03:00:00",
			"passed": "1 day"
		}
	]`

	timers, err := ParseSystemdTimersJSON(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(timers) != 1 {
		t.Fatalf("Expected 1 timer, got %d", len(timers))
	}

	timer := timers[0]
	if timer.Unit != "test.timer" {
		t.Errorf("Expected Unit 'test.timer', got '%s'", timer.Unit)
	}
	if timer.Service != "test.service" {
		t.Errorf("Expected Service 'test.service', got '%s'", timer.Service)
	}
	if timer.Next != "2024-01-15 03:00:00" {
		t.Errorf("Expected Next '2024-01-15 03:00:00', got '%s'", timer.Next)
	}
	if timer.Left != "10h" {
		t.Errorf("Expected Left '10h', got '%s'", timer.Left)
	}
}

func TestParseCrontabOutput_UserField(t *testing.T) {
	input := "0 * * * * /usr/bin/test.sh"
	jobs, _ := ParseCrontabOutput(input)
	
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}
	
	if jobs[0].User != "current" {
		t.Errorf("Expected User 'current', got '%s'", jobs[0].User)
	}
}

func TestParseCrontabOutput_FileField(t *testing.T) {
	input := "0 * * * * /usr/bin/test.sh"
	jobs, _ := ParseCrontabOutput(input)
	
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}
	
	if jobs[0].File != "user-crontab" {
		t.Errorf("Expected File 'user-crontab', got '%s'", jobs[0].File)
	}
}

func TestParseCrontabOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "trailing newline",
			input: "0 * * * * /usr/bin/test.sh\n",
			want:  1,
		},
		{
			name:  "multiple trailing newlines",
			input: "0 * * * * /usr/bin/test.sh\n\n\n",
			want:  1,
		},
		{
			name:  "leading whitespace",
			input: "  0 * * * * /usr/bin/test.sh",
			want:  1,
		},
		{
			name:  "mixed whitespace",
			input: "  \t 0 * * * * /usr/bin/test.sh  \t  ",
			want:  1,
		},
		{
			name:  "comment with leading space",
			input: "  # This is a comment\n0 * * * * /usr/bin/test.sh",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := ParseCrontabOutput(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(jobs) != tt.want {
				t.Errorf("Expected %d jobs, got %d", tt.want, len(jobs))
			}
		})
	}
}