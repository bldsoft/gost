package version

var Version, GitCommit, GitBranch string

func LongVersion() string {
	return Version + "#" + GitCommit
}
