package rules

// ResetCaches documents this exported function.
func ResetCaches() {
	clearParseFilesCache()
	clearTypedPackagesCache()
}
