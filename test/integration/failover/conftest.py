import pytest
from utils import run_command_and_get_output


# Fixtures
@pytest.fixture(scope="function")
def project_with_topology_and_vshard(cartridge_cmd, default_project_with_instances):
    project = default_project_with_instances.project

    cmd = [cartridge_cmd, "replicasets", "setup", "--bootstrap-vshard"]
    rc, _ = run_command_and_get_output(cmd, cwd=project.path)
    assert rc == 0

    return project
