// Package version provides the version of packages.
package version

// swagger:model VersionResponse
type Version struct {
	// The application build version
	// example: v1.0.0
	BuildVersion string `json:"buildVersion"`

	// The application build type
	// example: dev
	BuildType string `json:"buildType"`

	// The application build date
	// example: 2022-12-01
	BuildDate string `json:"buildDate"`
}

func DefaultVersion() *Version {
	return &Version{
		BuildType: "dev",
	}
}
