package rules

func ResetCaches() {
	clearParseFilesCache()
	clearTypedPackagesCache()
}
