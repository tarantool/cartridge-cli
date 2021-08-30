import copy
import os

import yaml
from utils import write_conf


def get_expelled_srv_conf(uuid):
    return {uuid: "expelled"}


def get_srv_conf(uuid, rpl_uuid, uri=None, disabled=False):
    return {
        uuid: {
            'disabled': disabled,
            'replicaset_uuid': rpl_uuid,
            'uri': uri if uri is not None else '%s-uri' % uuid,
        }
    }


def get_rpl_conf(uuid, leaders, alias=None):
    return {
        uuid: {
            'alias': alias if alias is not None else 'unnamed',
            'master': leaders,
            'roles': {'vshard-storage': True},
            'vshard_group': 'default',
            'weight': 1
        },
    }


def get_topology_conf(instances, replicasets):
    conf = {
        'failover': False,
        'replicasets': {},
        'servers': {},
    }

    for instance in instances:
        conf['servers'].update(instance)

    for replicaset in replicasets:
        conf['replicasets'].update(replicaset)

    return conf


def get_one_file_conf(instances, replicasets):
    return {
        'topology': get_topology_conf(instances, replicasets)
    }


class ClusterwideConfig:
    def __init__(self, conf, instance_uuid=None, replicaset_uuid=None, instance_uri=None, one_file=False):
        self.conf = conf
        self.instance_uuid = instance_uuid
        self.replicaset_uuid = replicaset_uuid
        self.instance_uri = instance_uri
        self.one_file = one_file


def write_instances_topology_conf(data_dir, app_name, conf, instances, one_file=False):
    conf_paths = []

    for instance in instances:
        work_dir = os.path.join(data_dir, '%s.%s' % (app_name, instance))
        os.makedirs(work_dir, exist_ok=True)

        if one_file:
            conf_path = os.path.join(work_dir, 'config.yml')
        else:
            conf_dir = os.path.join(work_dir, 'config')
            os.makedirs(conf_dir, exist_ok=True)

            conf_path = os.path.join(conf_dir, 'topology.yml')

        conf_paths.append(conf_path)
        write_conf(conf_path, conf)

    return conf_paths


def assert_conf_changed(conf_paths, other_app_conf_paths, old_conf, new_conf):
    for conf_path in conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == new_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert os.path.exists(backup_conf_path)

        with open(backup_conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

    if other_app_conf_paths is None:
        return

    # check that other app config wasn't changed
    for conf_path in other_app_conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert not os.path.exists(backup_conf_path)


def assert_conf_not_changed(conf_paths, old_conf):
    for conf_path in conf_paths:
        assert os.path.exists(conf_path)

        with open(conf_path, 'r') as f:
            conf = yaml.safe_load(f.read())
            assert conf == old_conf

        # check backup
        backup_conf_path = '%s.bak' % conf_path
        assert not os.path.exists(backup_conf_path)


def get_conf_with_new_uri(conf, instance_uuid, new_uri):
    new_conf = copy.deepcopy(conf)
    if new_conf.get('servers') is not None:
        new_conf['servers'][instance_uuid]['uri'] = new_uri
    else:
        new_conf['topology']['servers'][instance_uuid]['uri'] = new_uri

    return new_conf


def get_conf_with_removed_instance(conf, instance_uuid):
    new_conf = copy.deepcopy(conf)

    if new_conf.get('topology') is None:
        topology_conf = new_conf
    else:
        topology_conf = new_conf['topology']

    while True:
        if topology_conf['servers'][instance_uuid] == 'expelled':
            break

        replicaset_uuid = topology_conf['servers'][instance_uuid]['replicaset_uuid']

        # if there is no replicaset instance belong to - break
        if replicaset_uuid not in topology_conf['replicasets']:
            break

        # if instance not in replicaset leaders - break
        new_leaders = topology_conf['replicasets'][replicaset_uuid]['master']
        if isinstance(new_leaders, list) and instance_uuid not in new_leaders:
            break

        if isinstance(new_leaders, str) and instance_uuid != new_leaders:
            break

        rpl_other_instances = [
            uuid for uuid, instance_conf
            in topology_conf['servers'].items()
            if instance_conf.get('replicaset_uuid') == replicaset_uuid
            and uuid != instance_uuid
        ]
        rpl_other_instances.sort()

        # if instance was the last leader in replicaset, check if there are
        # other instances of this replicaset that aren't in the leaders list

        if isinstance(new_leaders, str):
            if len(rpl_other_instances) > 0:
                topology_conf['replicasets'][replicaset_uuid]['master'] = rpl_other_instances[0]
            else:
                del topology_conf['replicasets'][replicaset_uuid]

        if isinstance(new_leaders, list):
            new_leaders.remove(instance_uuid)

            if len(new_leaders) == 0:
                if len(rpl_other_instances) > 0:
                    new_leaders.append(rpl_other_instances[0])

            # leaders list is still empty - remove this replicaset
            if len(new_leaders) == 0:
                del topology_conf['replicasets'][replicaset_uuid]

        break

    del topology_conf['servers'][instance_uuid]

    return new_conf


def get_conf_with_new_leader(conf, replicaset_uuid, instance_uuid):
    new_conf = copy.deepcopy(conf)
    new_leaders = new_conf['replicasets'][replicaset_uuid]['master']
    if instance_uuid in new_leaders:
        new_leaders.remove(instance_uuid)

    new_leaders.insert(0, instance_uuid)

    return conf
