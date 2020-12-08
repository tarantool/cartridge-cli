import os

from utils import write_conf
from utils import get_instance_id
from utils import check_instances_running, check_instances_stopped
from utils import DEFAULT_CFG


class Instance():
    def __init__(self, name, http_port, advertise_uri):
        self.name = name
        self.http_port = http_port
        self.advertise_uri = advertise_uri

    def get_admin_api_url(self):
        return 'http://localhost:%s/admin/api' % self.http_port


class Replicaset():
    def __init__(self, name, instances):
        self.name = name
        self.instances = instances


class ProjectWithTopology():
    def __init__(self, cli, project, instances_list, replicasets_list=[], vshard_group_names=[]):
        self.cli = cli
        self.project = project
        self.instances = {i.name: i for i in instances_list}
        self.replicasets = {r.name: r for r in replicasets_list}
        self.vshard_group_names = vshard_group_names

        instances_conf = dict()
        for name, instance in self.instances.items():
            instances_conf.update({
                get_instance_id(project.name, name): {
                    'http_port': instance.http_port,
                    'advertise_uri': instance.advertise_uri,
                }
            })

        instances_conf_path = os.path.join(project.path, DEFAULT_CFG)
        if not os.path.exists(instances_conf_path):
            write_conf(instances_conf_path, instances_conf)

        self.instances_conf = instances_conf

    def set_replicasets(self, replicasets_list):
        self.replicasets = {r.name: r for r in replicasets_list}

    def start(self):
        self.cli.start(self.project, daemonized=True)
        check_instances_running(self.cli, self.project, [name for name in self.instances], daemonized=True)

    def stop(self):
        self.cli.stop(self.project, force=True)
        check_instances_stopped(self.cli, self.project, [name for name in self.instances])
        self.cli.clean(self.project)

        os.remove(os.path.join(self.project.path, DEFAULT_CFG))


def get_replicaset_by_alias(replicasets, alias):
    replicaset = None
    for r in replicasets:
        if r['alias'] == alias:
            replicaset = r
            break

    return replicaset


def get_list_from_log_lines(log_lines):
    return [
        line.split(maxsplit=1)[1]
        for line in log_lines
    ]
