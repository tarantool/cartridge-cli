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
