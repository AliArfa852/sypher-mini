// Package constants provides shared constants across the codebase.
package constants

// InternalChannels defines channels used for internal communication
// and should not be exposed to external users.
var InternalChannels = map[string]bool{
	"cli":    true,
	"system": true,
}

// IsInternalChannel returns true if the channel is internal.
func IsInternalChannel(channel string) bool {
	return InternalChannels[channel]
}
