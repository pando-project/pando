import os
import docker

HOST_PATH = '/opt/rancher/pando'
CONTAINER_PATH = '/root/.pando'
FILE = 'testfile.txt'

with open(os.path.join(HOST_PATH, FILE), 'w') as test_file:
    pass

docker_client = docker.from_env()
pando_container_id_stream = os.popen('docker ps | grep pando | awk \'{print $1}\'')
pando_container_id = pando_container_id_stream.read().strip()
pando_container = docker_client.containers.get(pando_container_id)
result = pando_container.exec_run('ls {container_file_path}'.format(
    container_file_path=os.path.join(CONTAINER_PATH, FILE)))
print(result)
