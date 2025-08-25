## alerteconso.com

### Environment variables

Those environment variables should be used when [building](#building) and [publishing](#publishing) the project

|  Variable name    |  Description                                                                         | Environment  |
| ----------------- | ------------------------------------------------------------------------------------ | ------------ |
| POSTGRES_USER     |  The PostgreSQL username                                                             | dev and prod |
| POSTGRES_PASSWORD | The PostgreSQL password                                                              | dev and prod |
| POSTGRES_HOST     | The PostgresSQL hostname                                                             | dev and prod |
| POSTGRES_DB       | The PostgreSQL database name                                                         | dev and prod |
| TAG               | The tag of the api image (default to `latest`)                                       | dev          |
| SSL_MODE          | The sslmode parameter used when connecting to database (only requird in production)  | prod         |

> [!TIP]
> You can create youself and `.env` file that defines values for those environment variables for `dev` and use `--env-file` option in the `docker compose` command instead of redfining each environment variable

### Local development

To run the application on your local machine (dev environment) please use following command to run the application in watch mode.

```bash
POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile dev watch
```

or if you have defined an `.env` file

```bash
docker compose --env-file <env_file_path> --profile dev watch
```

Keep your terminal open, the watch mode will follow the changes in the build context and update your application accordingly.

> [!NOTE]
> To clean up the image and shared volume properly use `docker compose down --volume`

### Publishing

Once you have finished the development publish your compose application by following the following steps:

> [!TIP]
> You can create youself and `.env` file that defines values for those environment variables for `prod` and use `--env-file` option in the `docker compose` command instead of redfining each environment variable

1. Build the images in production profile
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod build --platform linux/amd64,linux/arm64
   ```
   or if you have created an `.env` file
   ```bash
   docker compose --from-env <env_file_path> --profile prod build --platform linux/amd64,linux/arm64
   ```
2. Push the images to the local registry
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod push
   ```
    or if you have created an `.env` file
   ```bash
   docker compose --from-env <env_file_path> --profile prod push
   ```
3. Publish the Docker stack
   ```bash
   TAG=<tag_of_your_application> POSTGRES_USER=<postgresql_username> POSTGRES_PASSWORD=<postgresql_password> POSTGRES_HOST=<hostname_for_postgresql> docker compose --profile prod --resolve-image-digests publish <registry>/<repository>:<tag>
   ```
   or if you have created an `.env` file
   ```bash
   docker compose --from-env <env_file_path> --profile prod publish
   ```

### Deployment

First create an `.env` file for the production environment, that defines values for [environment variables](#environment-variables) if you did not already.

Once you have published the docker compose bundle and you have created the `.env` file, you can know deploy the application with: 
```bash
docker compose -f oci://<registry>/<repository>:<tag> --env-file <env_file_path> up -d
```
