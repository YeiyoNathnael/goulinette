package rules

// ResetCaches clears all package-level analysis caches maintained by the
// rules package (parsed-file cache and typed-packages cache). It must be
// called between analysis runs — typically at the start of Runner.Run —
// to prevent stale AST or type-info from a previous invocation from
// polluting the current run.
func ResetCaches() {
	clearParseFilesCache()
	clearTypedPackagesCache()
}
