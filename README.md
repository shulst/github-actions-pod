# github-actions-pod
Small Github Actions pod to run on private docker swarm

## Clusters
Two clusters, the clusters can swap roles or run both roles. The clusters are:
 - Snowflake
 - Waterdrop (Not yet active)

When both clusters are active, the main cluster will handle productive workloads. The secondary cluster will handle
staging workloads. If only 1 cluster is active, the cluster will handle all workloads.
The configuration of the different modes should use a different service storage location. Nginx will be try to send
staging workloads to the other cluster. It will add a header http_x_cluster_source = "slave" to the request, this way
the nginx can handle the request. If the header is missing the request will be send to the other cluster, unless the
other cluster is not active.

### Storage

Both clusters have a trueNAS storage running. They will be configured to sync with each other.
- nginx
  - Configuration for the nginx servers. Directory are snowflake (and waterdrop) and are mounted as /etc/nginx 
    on the nginx containers.
  - The nginx will run a service called `keepalived`, and should be configured to check if the containers are alive.
- projects
  - Storage for our Github repositories. The swarm manager has the permissions to read and write to github.
  - This can be used to build docker images, this repo (github-actions-pod) is also available there.
- service
  - Storage for the services running on the swarm. This allows for the docker pods to have persistent data.

### Snowflake

Snowflake is running on a proxmox server. It is accessible from the local network on 192.168.3.102.
It runs the following containers:
 - trueNAS
 - snowflake-swarm-manager
 - snowflake-swarm-worker
 - snowflake-nginx

#### Network

The containers are accessible from the following IP addresses:
- trueNAS
  - 192.168.3.111
  - 192.168.5.111
- snowflake-swarm-manager
  - 192.168.2.11
  - 192.168.5.11
- snowflake-swarm-worker
  - 192.168.2.20
  - 192.168.5.20
- snowflake-nginx
  - 192.168.2.21
  - 192.168.5.21


