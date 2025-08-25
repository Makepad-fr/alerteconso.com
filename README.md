## alerteconso.com

### Environment variables

Those environment variables should be used when [building](#building) and [publishing](#publishing) the project

|  Variable name    |  Description                                                                         |
| ----------------- | ------------------------------------------------------------------------------------ |
| POSTGRES_USER     |  The PostgreSQL username                                                             |
| POSTGRES_PASSWORD | The PostgreSQL password                                                              |
| POSTGRES_HOST     | The PostgresSQL hostname                                                             |
| POSTGRES_DB       | The PostgreSQL database name                                                         |
| TAG               | The tag of the api image (default to `latest`)                                       |
| SSL_MODE          | The sslmode parameter used when connecting to database (only requird in production)  |

### Building

To build the image on your local environment please use following command to run the application in watch mode.

```bash
POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile dev watch
```

Keep your terminal open, the watch mode will follow the changes in the build context and update your application accordingly.

> [!NOTE]
> To clean up the image and shared volume properly use `docker compose down --volume`

### Publishing

Once you have finished the development publish your compose application by following the following steps:

1. Build the images in production profile
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod build --platform linux/amd64,linux/arm64
   ```
2. Push the images to the local registry
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod push
   ```
3. Publish the Docker stack
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod --resolve-image-digests publish <registry>/<repository>:<tag>
   ```

### Deployment

First create an `.env` file for the production environment, that defines values for [environment variables](#environment-variables) if you did not already.

Once you have published the docker compose bundle and you have created the `.env` file, you can know deploy the application with `docker compose -f oci://<registry>/<repository>:<tag> --env-file <env_file_path> up -d`
