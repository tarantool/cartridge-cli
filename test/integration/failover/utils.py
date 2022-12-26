import requests
from utils import get_admin_url, get_response_data


def get_stateboard_failover_info():
    query = """
        query {
          cluster {
            failover_params {
                mode
                state_provider
                tarantool_params {
                    uri
                    password
                }
                failover_timeout
                fencing_enabled
                fencing_timeout
                fencing_pause
                autoreturn_delay
                leader_autoreturn
            }
          }
        }
    """

    response = requests.post(get_admin_url(8081), json={'query': query})
    return get_response_data(response)["cluster"]["failover_params"]


def get_common_failover_info():
    query = """
        query {
          cluster {
            failover_params {
                mode
                failover_timeout
                fencing_enabled
                fencing_timeout
                fencing_pause
            }
          }
        }
    """

    response = requests.post(get_admin_url(8081), json={'query': query})
    return get_response_data(response)["cluster"]["failover_params"]


def get_etcd2_failover_info():
    query = """
        query {
          cluster {
            failover_params {
                mode
                state_provider
                etcd2_params {
                    prefix
                    lock_delay
                    endpoints
                    username
                    password
                }
                failover_timeout
                fencing_enabled
                fencing_timeout
                fencing_pause
                autoreturn_delay
                leader_autoreturn
            }
          }
        }
    """

    response = requests.post(get_admin_url(8081), json={'query': query})
    return get_response_data(response)["cluster"]["failover_params"]


def assert_mode_and_params_state(failover_info, output):
    assert f"mode: {failover_info['mode']}" in output
    assert f"fencing_timeout: {failover_info['fencing_timeout']}" in output
    assert f"fencing_enabled: {failover_info['fencing_enabled']}".lower() in output
    assert f"fencing_pause: {failover_info['fencing_pause']}" in output
    assert f"failover_timeout: {failover_info['failover_timeout']}" in output
