def _routeguide_test_impl(ctx):
    # Build test execution script
    ctx.actions.write(ctx.outputs.executable, """set -x # Print commands
set -e # Fail on error

export DATABASE_FILE={database_file}
export SERVER_PORT={server_port}
export RUST_BACKTRACE=1 # Print rust stack traces

# Start server and wait
{server} &
sleep 2

# Run client
{client}

# Print completion for log
echo '---- DONE ----'
    """.format(
        client = ctx.executable.client.short_path,
        server = ctx.executable.server.short_path,
        database_file = ctx.file.database.short_path,
        server_port = ctx.attr.port,
    ), is_executable = True)

    # Build runfiles and default provider
    runfiles = ctx.runfiles(
        files = [ctx.executable.client, ctx.executable.server, ctx.file.database],
    )
    runfiles = runfiles.merge(ctx.attr.client[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.attr.server[DefaultInfo].default_runfiles)

    return [DefaultInfo(
        runfiles = runfiles,
    )]


routeguide_test = rule(
    implementation = _routeguide_test_impl,
    attrs = {
        "client": attr.label(
            doc = "Client binary",
            executable = True,
            mandatory = True,
            cfg = "target",
        ),
        "server": attr.label(
            doc = "Server binary",
            executable = True,
            mandatory = True,
            cfg = "target",
        ),
        "database": attr.label(
            doc = "Path to the feature database json file",
            mandatory = True,
            allow_single_file = True,
        ),
        "port": attr.int(
            doc = "Port to use for the client/server communication (value for SERVER_PORT env var)",
            default = 50051,
        ),
    },
    test = True,
)


def get_parent_dirname(label):
    if label.startswith("//"):
        label = label[2:]
    return label.partition("/")[0]


def routeguide_test_matrix(clients = [], servers = [], database = "//example/proto:routeguide_features", tagmap = {}):
    """Build a matrix of tests that checks every client against every server"""
    port = 50051
    for server in servers:
        server_name = get_parent_dirname(server)
        for client in clients:
            client_name = get_parent_dirname(client)
            name = "%s_%s" % (client_name, server_name)

            # Extract tags for client and server
            tags = []
            if tagmap.get(client_name):
                tags.extend(tagmap.get(client_name))
            if tagmap.get(server_name):
                tags.extend(tagmap.get(server_name))
            if tagmap.get(name):
                tags.extend(tagmap.get(name))

            # Setup test with next available port number
            routeguide_test(
                name = name,
                client = client,
                server = server,
                database = database,
                port = port,
                tags = tags,
                size = "small",
            )

            port += 1
