import docker

client=docker.from_env()

containers = client.containers.list()

for container in containers:
    try:
        processes=container.top(ps_args="aux")
        print(processes["Processes"])
    except docker.errors.APIError:
        pass
print(containers)
