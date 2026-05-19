package scheduler

import (
	"os"
	"os/exec"
	"testing"
)

func TestParseCrontabFile(t *testing.T) {
	// Create a temporary crontab file
	tmpFile, err := os.CreateTemp("", "crontab-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `# Test crontab
0 5 * * * /usr/bin/backup.sh
*/15 * * * * /usr/bin/monitor.sh
@reboot /usr/bin/startup.sh`

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Read and parse the file
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	jobs, err := ParseCrontabOutput(string(data))
	if err != nil {
		t.Fatalf("ParseCrontabOutput failed: %v", err)
	}

	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}
}

func TestCrontabRoundTrip(t *testing.T) {
	original := "0 5 * * * /usr/bin/backup.sh"
	jobs, err := ParseCrontabOutput(original)
	if err != nil {
		t.Fatalf("ParseCrontabOutput failed: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job, got %d", len(jobs))
	}

	if jobs[0].Schedule != "0 5 * * *" {
		t.Errorf("Expected schedule '0 5 * * *', got '%s'", jobs[0].Schedule)
	}

	if jobs[0].Command != "/usr/bin/backup.sh" {
		t.Errorf("Expected command '/usr/bin/backup.sh', got '%s'", jobs[0].Command)
	}
}

func TestSystemdTimersReal(t *testing.T) {
	// Try to get actual systemd timers if available
	cmd := exec.Command("systemctl", "list-timers", "--output=json", "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skip("systemctl not available or no timers found")
		return
	}

	timers, err := ParseSystemdTimersJSON(string(output))
	if err != nil {
		t.Fatalf("ParseSystemdTimersJSON failed: %v", err)
	}

	t.Logf("Found %d systemd timers", len(timers))
	// Just verify it doesn't crash
}

func TestCrontabWithEnvironment(t *testing.T) {
	input := `SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

0 5 * * * /usr/bin/backup.sh`

	jobs, err := ParseCrontabOutput(input)
	if err != nil {
		t.Fatalf("ParseCrontabOutput failed: %v", err)
	}

	// Environment variables should be skipped
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job (environment vars skipped), got %d", len(jobs))
	}
}

func TestCrontabComplexCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantJobs int
	}{
		{
			name:     "command with pipes",
			input:    "0 * * * * /usr/bin/script.sh | /usr/bin/other.sh",
			wantJobs: 1,
		},
		{
			name:     "command with redirection",
			input:    "0 * * * * /usr/bin/script.sh > /var/log/script.log 2>&1",
			wantJobs: 1,
		},
		{
			name:     "command with variables",
			input:    "0 * * * * /usr/bin/script.sh $HOME /tmp",
			wantJobs: 1,
		},
		{
			name:     "command with quotes",
			input:    "0 * * * * /usr/bin/script.sh \"arg with spaces\"",
			wantJobs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs, err := ParseCrontabOutput(tt.input)
			if err != nil {
				t.Fatalf("ParseCrontabOutput failed: %v", err)
			}
			if len(jobs) != tt.wantJobs {
				t.Errorf("Expected %d jobs, got %d", tt.wantJobs, len(jobs))
			}
		})
	}
}