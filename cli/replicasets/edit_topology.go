package replicasets

import (
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/fatih/structs"

	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

type JoinInstanceOpts struct {
	URI string `structs:"uri"`
}

type EditReplicasetOpts struct {
	ReplicasetUUID  string `structs:"uuid,omitempty"`
	ReplicasetAlias string `structs:"alias,omitempty"`

	Roles       []string `structs:"roles,omitempty"`
	AllRW       *bool    `structs:"all_rw,omitempty"`
	Weight      *float64 `structs:"weight,omitempty"`
	VshardGroup *string  `structs:"vshard_group,omitempty"`

	JoinInstances         []JoinInstanceOpts `structs:"join_servers,omitempty"`
	FailoverPriorityUUIDs []string           `structs:"failover_priority,omitempty"`
}

type EditReplicasetsListOpts []*EditReplicasetOpts

type EditInstanceOpts struct {
	InstanceUUID string `structs:"uuid,omitempty"`
	Expelled     bool   `structs:"expelled,omitempty"`
}

type EditInstancesListOpts []*EditInstanceOpts

func (listOpts *EditReplicasetsListOpts) ToMapsList() []map[string]interface{} {
	optsMapsList := make([]map[string]interface{}, len(*listOpts))
	for i, opts := range *listOpts {
		optsMapsList[i] = structs.Map(opts)
	}

	return optsMapsList
}

func (listOpts *EditInstancesListOpts) ToMapsList() []map[string]interface{} {
	optsMapsList := make([]map[string]interface{}, len(*listOpts))
	for i, opts := range *listOpts {
		optsMapsList[i] = structs.Map(opts)
	}

	return optsMapsList
}

var (
	editReplicasetsBody string
)

func init() {
	var err error

	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		panic(fmt.Errorf("Failed to compute get topology replicaset function body: %s", err))
	}

	editReplicasetsBody, err = templates.GetTemplatedStr(&editReplicasetsBodyTemplate, map[string]string{
		"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
	})

	if err != nil {
		panic(fmt.Errorf("Failed to compute edit_topology call body: %s", err))
	}
}

func editReplicasetsList(conn *connector.Conn, opts *EditReplicasetsListOpts) (*TopologyReplicasets, error) {
	waitForHealthy, err := healthCheckIsNeeded(conn)
	if err != nil {
		return nil, err
	}

	req := connector.EvalReq(editReplicasetsBody, opts.ToMapsList())

	var newTopologyReplicasetsList []*TopologyReplicaset
	if err := conn.ExecTyped(req, &newTopologyReplicasetsList); err != nil {
		return nil, fmt.Errorf("Failed to edit topology: %s", err)
	}

	newTopologyReplicasets := getTopologyReplicasetsFromList(newTopologyReplicasetsList)

	if waitForHealthy {
		if err := waitForClusterIsHealthy(conn); err != nil {
			return nil, fmt.Errorf("Failed to wait for cluster to become healthy: %s", err)
		}
	}

	return newTopologyReplicasets, nil
}

func editReplicaset(conn *connector.Conn, opts *EditReplicasetOpts) (*TopologyReplicaset, error) {
	editReplicasetsOpts := &EditReplicasetsListOpts{opts}
	newTopologyReplicasets, err := editReplicasetsList(conn, editReplicasetsOpts)
	if err != nil {
		return nil, err
	}

	if len(*newTopologyReplicasets) != 1 {
		return nil, project.InternalError("One replicaset should be returned, got %#v", newTopologyReplicasets)
	}

	newTopologyReplicaset := newTopologyReplicasets.GetSomeReplicaset()

	return newTopologyReplicaset, nil
}

func editInstances(conn *connector.Conn, opts *EditInstancesListOpts) (bool, error) {
	req := connector.EvalReq(editInstanceBody, opts.ToMapsList())

	if _, err := conn.Exec(req); err != nil {
		return false, fmt.Errorf("Failed to edit topology: %s", err)
	}

	return true, nil
}

func waitForClusterIsHealthy(conn *connector.Conn) error {
	retryOpts := []retry.Option{
		retry.MaxDelay(1 * time.Second),
		retry.Attempts(30),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return !strings.Contains(err.Error(), "Received in bad format")
		}),
	}

	checkClusterIsHealthyFunc := func() error {
		req := connector.EvalReq(getClusterIsHealthyBody)
		var isHealthy bool

		if err := conn.ExecTyped(req, &isHealthy); err != nil {
			return fmt.Errorf("Failed to check cluster is healthy: %s", err)
		}

		if !isHealthy {
			return fmt.Errorf("Cluster isn't healthy")
		}

		return nil
	}

	return retry.Do(checkClusterIsHealthyFunc, retryOpts...)
}

var (
	tableTemplate = `{ {{ .OptsString }} }`
)
