package cmd

const (
	ExitOK         = 0
	ExitFindings   = 1 // Findings found above threshold
	ExitGeneric    = 2
	ExitUsage      = 3
	ExitFilesystem = 4
	ExitNotGitRepo = 5
)
