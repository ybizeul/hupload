![Icon](readme_images/icon.svg#gh-light-mode-only)
![Icon](readme_images/icon-dark.svg#gh-dark-mode-only)

# Hupload

Hupload is a minimalist file uploader for your customers.

With technical support in mind, it simplifies the process of receiving log files
or support bundles from your users.

It is a web portal and an API for direct integration into your products.

The overall concept is that share names are random tokens that can be generated
by **Hupload** and are publicly accessible so users don't have to log in to 
upload content.

Note that "Copy" buttons will not work over insecure connection (http without 
SSL)

## Screenshots

**Admin page**

![Shares list](readme_images/shares-dark.png#gh-dark-mode-only)
![Shares list](readme_images/shares-light.png#gh-light-mode-only)

**Share page with upload box**

![Share content](readme_images/share-dark.png#gh-dark-mode-only)
![Share content](readme_images/share-light.png#gh-light-mode-only)

**Advanced share settings**

![Share properties](readme_images/properties-dark.png#gh-dark-mode-only)
![Share properties](readme_images/properties-light.png#gh-light-mode-only)

**Markdown preview**

![Markdown preview](readme_images/properties-preview-dark.png#gh-dark-mode-only)
![Markdown preview](readme_images/properties-preview-light.png#gh-light-mode-only)

## Environment variables

| Variable     | Description |
|--------------|-------------|
| `CONFIG`     | Path to `config.yml`    |
| `HTTP_PORT`  | Port to run web service |
| `JWT_SECRET` | Random string used to sign sessions cookies |

## Features

- Quickly create random links and share with users,
- Easy to use drag and drop interface for uploads,
- S3 or filesystem storage,
- Configurable max share size and max file size,
- Basic share informations listed (number of items, total size),
- Add instructions in Markdown for your users and define your own reusable templates.
- Automatic dark mode following OS settings,
- Multi user (all admins see all shares, but see their own listed separately first),
- Flat user file or OIDC authentication,
- API first, everything can be done through REST calls,
- Minimalist, clean interface.

## Configuration

By default, Hupload uses `data/` directory and `admin` user with a randomly
generated password that is displayed in the logs during startup.

If a configuration file is provided in `CONFIG` environment variable, it will
be used to configure storage backend and authentication.

Sample configuration file :

```
Title: Hupload
default_exposure: upload
default_validity_days: 7
auth:
  type: file
  options:
    path: config/users.yml
storage:
  type: file
  options:
    path: data
    max_file_mb: 512
    max_share_mb: 2048
```

For authentication, users are defined in a yaml file :

```
- username: user1
  password: $2y$10$LIcTF3HKNhV6qh3oi3ysHOnhiXpLOU22N61JzZXoSWQbNOpDhS/g.
- username: user2
  password: $2y$10$Rwj3rjfmXuflxds.uhgKReiXFy5VRziYuDDw/aO1w9ut9BzafTFr6
```

Passwords are hashes that you can generate with `htpasswd` :

To generate a hash for `hupload` password string :

```
htpasswd -bnBC 10 "" hupload | tr -d ":"
$2y$10$LIcTF3HKNhV6qh3oi3ysHOnhiXpLOU22N61JzZXoSWQbNOpDhS/g.
```

### S3 Storage

```
storage:
  type: s3
  options:
    region: us-east-1
    endpoint: my.s3server.com
    use_path_style: true
    aws_key: <aws_key>
    aws_secret: <aws_secret>
    bucket: hupload
    max_file_mb: 512
    max_share_mb: 2048
```

S3 options can also be set in environment using `AWS_DEFAULT_REGION`,
`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_ENDPOINT_URL`. Options set
in configuration file have precedence.

Note that `region` is mandatory for AWS API to work correctly even if you
are using your own S3 server like minio.

### OIDC 

OIDC redirect url is `/oidc` and you can provide application details with the
following configuration :

```
auth:
  type: oidc
  options:
    provider_url: https://auth.tynsoe.org/application/o/hupload/
    client_id: <client_id>
    client_secret: <client_secret>
    redirect_url: https://hupload.company.com/oidc
```

### Canned messages

When creating a share you can use markdown to display a custom message with
instructions to the user.

You can pre-define canned messages in the configuration file :

```
messages:
  - title: Welcome to Hupload
    message: |
      Hupload is a simple file sharing service.

      You can upload files up to 3MB and share up to 5MB.

      Enjoy!
  - title: Log collection
    message: |
      ### Collecting support bundle

      You can download a support bundle from the home page in the
      **About** menu. Download the file and upload it here.
```

## Run in a container

You can quickly test **Hupload** in a container, or run it in production :

Note that the container is built to run _Hupload_ as **nonroot** user (65532),
and needs access to the shares directory

```
mkdir data
chown 65532:65532 data
docker run -v $(pwd)/data:/data -p 8080:8080 ghcr.io/ybizeul/hupload/hupload
```

Alternatively, you can use the `compose.yml` file provided. Username is `admin`
and password is `hupload` as defined in `hupload/config/users.yml.sample`

```
docker compose up
```
## API

The following endpoints are available under `/api/v1`

**Basic Authentication Required**

| Type     | URL                            | Description                          |
|----------|--------------------------------|--------------------------------------|
| `GET`    | `/shares`                      | Get a list of all shares
| `POST`   | `/shares`                      | Create a new share with a random name (See parameters)
| `POST`   | `/shares/{share}`              | Create a new share named `{share}` (See parameters)
| `PATCH`  | `/shares/{share}`              | Update share parameters (See parameters)
| `DELETE` | `/shares/{share}`              | Delete a share and all its content
| `GET`    | `/shares/{share}/items/{item}` | Get an `{item}` (file) content. Authentication not required if share is exposed as `download` or `both`
| `GET`    | `/d/{share}/{item}` | Alias to get an file content (See above)

**Public Endpoints**

| Type     | URL                            | Description                          |
|----------|--------------------------------|--------------------------------------|
| `POST`   | `/shares/{share}/items/{item}` | Post a new file `{item}` in `{share}` (multipart form encoded)
| `GET`    | `/shares/{share}`              | Get a `{share}` content

**Parameters**

When creating or updateing a new share, you can define parameters in the JSON body :

| Key           | Type                               | Description                          |
|------------   |------------------------------------|--------------------------------------|
| `validity`    | `number`                           | Number of days the share is valid
| `exposure`    | `enum["upload","download","both"]` | Whether guest users can upload files, download files or do both
| `description` | `string`                           | A short description displayed in shares view
| `message`     | `url encoded markdown`             | Instructions in markdown visible to the guest
