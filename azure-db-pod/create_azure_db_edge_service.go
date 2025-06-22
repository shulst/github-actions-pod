package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run script.go <sa_password> <external_port>")
		os.Exit(1)
	}

	saPassword := os.Args[1]
	externalPort, err := strconv.ParseUint(os.Args[2], 10, 16)
	if err != nil {
		fmt.Println("Error: External port must be a valid number")
		os.Exit(1)
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Function to create the Azure DB Edge service
	err = createAzureDBEdgeService(ctx, cli, saPassword, uint16(externalPort))
	if err != nil {
		panic(err)
	}

	fmt.Println("Azure DB Edge service created successfully")
}

func createAzureDBEdgeService(ctx context.Context, cli *client.Client, saPassword string, externalPort uint16) error {
	serviceName := "azure-db-edge"
	imageName := "mcr.microsoft.com/mssql/server:2022-latest"

	mounts := []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: "azure-db-edge",
			Target: "/var/opt/mssql",
			VolumeOptions: &mount.VolumeOptions{
				DriverConfig: &mount.Driver{
					Name: "local",
					Options: map[string]string{
						"type":   "nfs",
						"o":      "addr=192.168.2.42,rw,vers=3",
						"device": ":/swarm01-data/service/azure-db-edge/",
					},
				},
			},
		},
		// New mount for the certificate
		//{
		//	Type:     mount.TypeBind,
		//	Source:   "/mnt/swarm-data/environments/development/certs/dev.cert",
		//	Target:   "/var/opt/mssql/certs/dev.cert",
		//	ReadOnly: true,
		//},
		//{
		//	Type:     mount.TypeBind,
		//	Source:   "/mnt/swarm-data/environments/development/certs/dev.key",
		//	Target:   "/var/opt/mssql/certs/dev.key",
		//	ReadOnly: true,
		//},
	}

	serviceSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: serviceName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:  imageName,
				Mounts: mounts,
				Env: []string{
					"ACCEPT_EULA=Y",
					fmt.Sprintf("MSSQL_SA_PASSWORD=%s", saPassword),
					"MSSQL_PID=Developer",
					"MSSQL_MEMORY_LIMIT_MB=4096",
					// New environment variables for SSL configuration
					//"MSSQL_SSL_CERT=/var/opt/mssql/certs/sqlserver.cert",
					//"MSSQL_SSL_KEY=/var/opt/mssql/certs/sqlserver.key",
					//"MSSQL_SSL_CA=/var/opt/mssql/certs/dev.cert",
					"MSSQL_ENCRYPT=OPTIONAL",
				},
				CapabilityAdd: []string{"CAP_SYS_ADMIN"},
				//Command:       []string{"/opt/mssql/bin/sqlservr", "-c", "-f", "--accept-eula", "--force-ssl"},
			},
		},
		//Networks: []swarm.NetworkAttachmentConfig{
		//	{
		//		Target: "host",
		//	},
		//},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &[]uint64{1}[0],
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      swarm.PortConfigProtocolTCP,
					PublishedPort: uint32(externalPort),
					TargetPort:    1433,
				},
			},
		},
	}

	_, err := cli.ServiceCreate(ctx, serviceSpec, types.ServiceCreateOptions{})
	return err
}
