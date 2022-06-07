package replicasets

import _ "embed"

//go:embed lua/bootstrap_vshard_body.lua
var bootstrapVshardBody string

//go:embed lua/get_cluster_is_healthy_body.lua
var getClusterIsHealthyBody string

//go:embed lua/edit_instance_body.lua
var editInstanceBody string

//go:embed lua/edit_replicasets_body_template.lua
var editReplicasetsBodyTemplate string

//go:embed lua/format_topology_replicaset_func_template.lua
var formatTopologyReplicasetFuncTemplate string

//go:embed lua/get_known_roles_body.lua
var getKnownRolesBody string

//go:embed lua/get_known_vshard_groups_body.lua
var getKnownVshardGroupsBody string

//go:embed lua/get_topology_replicasets_body_template.lua
var getTopologyReplicasetsBodyTemplate string
