package version

var (
	Version = "dev"
	Commit  = "unknown"
)

func String() string {
	if Commit == "" || Commit == "unknown" {
		return Version
	}
	return Version + " (" + Commit + ")"
}
