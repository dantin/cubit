package version

import "fmt"

// ApplicationVersion represents application version.
var ApplicationVersion = NewVersion(0, 0, 1)

// SematicVersion represents version information with Sematic Versioning specification.
type SematicVersion struct {
	major uint
	minor uint
	patch uint
}

// NewVersion initializes a new instance of SematicVersion.
func NewVersion(major, minor, patch uint) *SematicVersion {
	return &SematicVersion{
		major: major,
		minor: minor,
		patch: patch,
	}
}

// String returns a string that represents this instance.
func (v *SematicVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

// IsEqual returns true if version instance is equal to the second.
func (v *SematicVersion) IsEqual(v2 *SematicVersion) bool {
	if v == v2 {
		return true
	}
	return v.major == v2.major && v.minor == v2.minor && v.patch == v2.patch
}

// IsLess returns true if version instance is less than the second.
func (v *SematicVersion) IsLess(v2 *SematicVersion) bool {
	if v == v2 {
		return false
	}
	if v.major == v2.major {
		if v.minor == v2.minor {
			if v.patch == v2.patch {
				return false
			}
			return v.patch < v2.patch
		}
		return v.minor < v2.minor
	}
	return v.major < v2.major
}

// IsLessOrEqual returns true if version instance is less than or equal to the second.
func (v *SematicVersion) IsLessOrEqual(v2 *SematicVersion) bool {
	return v.IsLess(v2) || v.IsEqual(v2)
}

// IsGreater returns true if version instance is greater than the second.
func (v *SematicVersion) IsGreater(v2 *SematicVersion) bool {
	if v == v2 {
		return false
	}
	if v.major == v2.major {
		if v.minor == v2.minor {
			if v.patch == v2.patch {
				return false
			}
			return v.patch > v2.patch
		}
		return v.minor > v2.minor
	}
	return v.major > v2.major
}

// IsGreaterOrEqual returns true if version instance is greater than or equal to the second.
func (v *SematicVersion) IsGreaterOrEqual(v2 *SematicVersion) bool {
	return v.IsGreater(v2) || v.IsEqual(v2)
}
