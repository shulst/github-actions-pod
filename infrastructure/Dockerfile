version: '3.8'

services:
  db:
    image: mcr.microsoft.com/azure-sql-edge:latest
    environment:
      - ACCEPT_EULA=1
      - MSSQL_SA_PASSWORD=${DB_PASSWORD}
    ports:
      - "${DB_PORT}:1433"
    volumes:
      - type: bind
        source: /swarm01-data/services/${SERVICE_DIR}
        target: /var/opt/mssql
    deploy:
      placement:
        constraints:
          - node.role == worker

  web:
    image: ${WEB_IMAGE}
    ports:
      - "${WEB_PORT}:80"
    volumes:
      - type: bind
        source: /swarm01-data/services/${SERVICE_DIR}
        target: /app/data
    deploy:
      placement:
        constraints:
          - node.role == worker
