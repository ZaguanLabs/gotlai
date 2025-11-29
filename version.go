package gotlai

// Version information for gotlai.
// These values can be overridden at build time using ldflags:
//
//	go build -ldflags "-X github.com/ZaguanLabs/gotlai.Version=1.0.0"
const (
	// Name is the application name.
	Name = "gotlai"

	// Description is a short description of the application.
	Description = "Go Translation AI - AI-powered HTML translation engine"

	// Version is the semantic version of the application.
	// Override at build time with ldflags for releases.
	Version = "0.1.0"

	// Repository is the source code repository URL.
	Repository = "https://github.com/ZaguanLabs/gotlai"

	// License is the software license.
	License = "MIT"
)

// BuildInfo contains build-time information.
// These are typically set via ldflags during build.
var (
	// GitCommit is the git commit hash.
	GitCommit = "unknown"

	// GitBranch is the git branch name.
	GitBranch = "unknown"

	// BuildDate is the build timestamp.
	BuildDate = "unknown"

	// GoVersion is the Go version used to build.
	GoVersion = "unknown"
)

// FullVersion returns the version string with optional build info.
func FullVersion() string {
	v := Version
	if GitCommit != "unknown" && GitCommit != "" {
		short := GitCommit
		if len(short) > 7 {
			short = short[:7]
		}
		v += "+" + short
	}
	return v
}

// UserAgent returns a user agent string for HTTP requests.
func UserAgent() string {
	return Name + "/" + Version
}
