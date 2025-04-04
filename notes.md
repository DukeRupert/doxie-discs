## To get ip address of psql container on codespaces
```bash 
DB_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' devcontainer-db-1)
```