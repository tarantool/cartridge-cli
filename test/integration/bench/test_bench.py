import os
import signal
import socket
from subprocess import PIPE, STDOUT, Popen
from threading import Thread

import tenacity
from utils import consume_lines, run_command_and_get_output


@tenacity.retry(stop=tenacity.stop_after_delay(15), wait=tenacity.wait_fixed(1))
def wait_for_connect():
    socket.create_connection(('127.0.0.1', 3301))


def test_bench(cartridge_cmd, request, tmpdir):
    base_cmd = [cartridge_cmd, 'bench', '--duration=1']
    tarantool_cmd = [
        "tarantool",
        "-e", f"""box.cfg{{listen="127.0.0.1:3301",work_dir=[[{tmpdir}]]}}""",
        "-e", """box.schema.user.grant("guest","super",nil,nil,{if_not_exists=true})"""
    ]

    env = os.environ.copy()
    process = Popen(tarantool_cmd, stdout=PIPE, stderr=STDOUT, env=env)
    thread = Thread(target=consume_lines, args=["3301", process.stdout])
    thread.start()

    def kill():
        process.send_signal(signal.SIGKILL)
        if thread is not None:
            thread.join(5)
    request.addfinalizer(kill)

    wait_for_connect()

    rc, output = run_command_and_get_output(base_cmd, cwd=tmpdir)
    assert rc == 0

    base_cmd = [cartridge_cmd, 'bench', '--duration=1', '--fill=1000']
    rc, output = run_command_and_get_output(base_cmd, cwd=tmpdir)
    assert rc == 0

    base_cmd = [cartridge_cmd, 'bench', '--duration=1', '--insert=0', '--select=50', '--update=50']
    rc, output = run_command_and_get_output(base_cmd, cwd=tmpdir)
    assert rc == 0
