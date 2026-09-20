package dashboard

// Commands wraps Service for mutating operations: prune, pull, and sync/views config.
// Read operations (Collect, ClearCaches) remain on Service.
type Commands struct {
	*Service
}

// NewCommands returns a Commands backed by the given Service.
func NewCommands(s *Service) *Commands {
	return &Commands{Service: s}
}
