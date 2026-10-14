package gitstore

import (
	"github.com/go-git/go-git/v6/plumbing/client"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
)

// Auth carries credentials for clone/push against a remote forge (used by
// GitHub mirroring). It is intentionally minimal; extend as mirror needs grow.
type Auth struct {
	// Username/Password for HTTPS basic auth (e.g. a GitHub PAT as Password).
	Username string
	Password string
	// PEM is an SSH private key (PEM bytes) for SSH transport.
	PEM []byte
}

// clientOptions converts Auth to go-git v6 client.Option values, or nil if
// unset. go-git v6 configures transport auth via ClientOptions rather than an
// Auth field on CloneOptions/PushOptions.
func (a *Auth) clientOptions() []client.Option {
	if a == nil {
		return nil
	}
	if a.Username != "" || a.Password != "" {
		return []client.Option{client.WithHTTPAuth(&githttp.BasicAuth{
			Username: a.Username,
			Password: a.Password,
		})}
	}
	if len(a.PEM) > 0 {
		// SSH auth requires the ssh client config; mirror SSH support is added
		// when Phase 5 mirror needs it. For now HTTPS is the supported path.
		return nil
	}
	return nil
}
