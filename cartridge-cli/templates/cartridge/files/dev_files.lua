local dev_files = {
    {
        name = 'deps.sh',
        mode = tonumber('0755', 8),
        content = [=[
            #!/bin/sh
            # Call this scripts to install test dependencies

            set -e

            # Test dependencies:
            tarantoolctl rocks install luatest 0.5.0
            tarantoolctl rocks install luacov 0.13.0
            tarantoolctl rocks install luacheck 0.25.0
        ]=]
    },
    {
        name = 'instances.yml',
        mode = tonumber('0644', 8),
        content = [=[
            ${project_name_lower}.router:
            workdir: ./tmp/db_dev/3301
            advertise_uri: localhost:3301
            http_port: 8081

          ${project_name_lower}.s1-master:
            workdir: ./tmp/db_dev/3302
            advertise_uri: localhost:3302
            http_port: 8082

          ${project_name_lower}.s1-replica:
            workdir: ./tmp/db_dev/3303
            advertise_uri: localhost:3303
            http_port: 8083

          ${project_name_lower}.s2-master:
            workdir: ./tmp/db_dev/3304
            advertise_uri: localhost:3304
            http_port: 8084

          ${project_name_lower}.s2-replica:
            workdir: ./tmp/db_dev/3305
            advertise_uri: localhost:3305
            http_port: 8085
        ]=]
    },
    {
        name = '.cartridge.yml',
        mode = tonumber('0644', 8),
        content = [=[
            ---
            run_dir: 'tmp'
        ]=]
    },
    {
        name = 'tmp/.keep',
        mode = tonumber('0644', 8),
        content = '',
    }
}

return dev_files
