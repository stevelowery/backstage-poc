module gitlab.mgmt.arms-dev.net/${{ values.namespace }}/ ${{ values.component_id }}

go 1.17

require (
	gitlab.mgmt.arms-dev.net/go-common/healthcheck v0.0.0-20200415152312-00b2966a1616
	gitlab.mgmt.arms-dev.net/go-common/logger v0.0.7
)

require (
	github.com/sirupsen/logrus v1.8.1 // indirect
	gitlab.mgmt.arms-dev.net/go-common/accesslog v0.0.0-20200405190917-d2288d205e68 // indirect
	golang.org/x/sys v0.0.0-20200331124033-c3d80250170d // indirect
)
