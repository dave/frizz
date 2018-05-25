package config

import (
	"time"
)

const (
	DevServerPort          = 8090
	LocalFileserverTempDir = "/Users/dave/.frizz-local"

	// ProjectId is the ID of the GCS project
	ProjectID = "TODO"

	// FrizzHost is the domain
	FrizzHost = "frizz.io"

	// MaxConcurrentRequests is the maximum number of concurrent requests per server
	MaxConcurrentRequests = 2

	// MaxQueue is the maximum queue length waiting for compile. After this an error is returned.
	MaxQueue = 100

	AssetsFilename = "assets.zip"

	// WriteTimeout is the timeout when serving static files
	WriteTimeout = time.Second * 2

	// PageTimeout is the timeout when generating the compile page
	PageTimeout = time.Second * 5

	// ServerShutdownTimeout is the timeout when doing a graceful server shutdown
	ServerShutdownTimeout = time.Second * 5

	// WebsocketPingPeriod is the interval between pings. Must be less than WebsocketPongTimeout.
	WebsocketPingPeriod = time.Second * 10

	// WebsocketPongTimeout is the time to wait for a pong from the client before cancelling
	WebsocketPongTimeout = time.Second * 20

	// WebsocketWriteTimeout is the write timeout for websockets
	WebsocketWriteTimeout = time.Second * 20

	// WebsocketInstructionTimeout is the time to wait for instructions from the client (e.g. during
	// playground compile)
	WebsocketInstructionTimeout = time.Second * 5

	// WebsocketTimeout is the overall timeout for websockets
	WebsocketTimeout = time.Second * 360

	// GitSaveTimeout is the timeout when saving git repos to GCS
	GitSaveTimeout = time.Second * 300

	// GitCloneTimeout is the time to wait for a git clone operation
	GitCloneTimeout = time.Second * 120

	// GitPullTimeout is the time to wait for a git pull operation
	GitPullTimeout = time.Second * 60

	// GitListTimeout is the time to wait for a git list operation
	GitListTimeout = time.Second * 10

	// GitMaxObjects is the maximum objects in git clone progress
	GitMaxObjects = 30000

	// HttpTimeout is the time to wait for HTTP operations (e.g. getting meta data - not git)
	HttpTimeout = time.Second * 5

	ConcurrentStorageUploads = 10
)

var ValidExtensions = []string{".go", ".jsgo.html", ".inc.js", ".md"}

var Buckets = []string{PkgBucket, GitBucket}

var Sites = map[string]string{
	PkgHost: PkgBucket,
}
