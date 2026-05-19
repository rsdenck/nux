// Package skill provides skill management for NUX CLI.
//
// Skills are external CLI tool integrations defined in .md files.
// Each skill can be installed, enabled, and managed.
//
// This package includes:
//   - Skill: Represents an external CLI tool integration
//   - Vault: Secure storage for skill configuration and API keys
//   - LoadSkillFromMD: Load skill from markdown file
//   - ListSkills: List all available skills
//   - SearchSkills: Search skills by name or type
//
// Vault features:
//   - Encrypted storage with AES-256-GCM
//   - Argon2id key derivation
//   - Optional passphrase protection
//   - Backward compatible with plaintext
//
// Example usage:
//
//	// Load a skill
//	skill, err := skill.LoadSkillFromMD("docker")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Install the skill
//	if err := skill.Install(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load vault with optional encryption
//	v, err := skill.LoadVault("my-passphrase")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save vault with encryption
//	if err := skill.SaveVault(v, "my-passphrase"); err != nil {
//	    log.Fatal(err)
//	}
package skill