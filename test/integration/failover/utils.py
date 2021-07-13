import requests

from utils import (
    get_admin_url,
    get_response_data
)


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
            }
          }
        }
    """

    response = requests.post(get_admin_url(8081), json={'query': query})
    return get_response_data(response)["cluster"]["failover_params"]


def get_eventual_failover_info():
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
            }
          }
        }
    """

    response = requests.post(get_admin_url(8081), json={'query': query})
    return get_response_data(response)["cluster"]["failover_params"]
