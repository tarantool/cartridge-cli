import requests


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


def set_instance_zone(admin_api_url, instance_name, zone):
    # get instance UUID
    query = '''
        query {
            servers: servers {
                uuid
                alias
            }
        }
    '''

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
    instances = resp['data']['servers']

    uuid = None
    for instance in instances:
        if instance['alias'] == instance_name:
            uuid = instance['uuid']
            break

    assert uuid is not None

    query = '''
        mutation {
        set_zone: cluster{
            edit_topology(
                servers: [{
                    uuid: "%s",
                    zone: "%s",
                }]
            ) { }
        }}
        ''' % (uuid, zone)

    r = requests.post(admin_api_url, json={'query': query})
    assert r.status_code == 200
    resp = r.json()
    assert 'data' in resp
