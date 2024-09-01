![Icon](readme_images/icon.svg#gh-light-mode-only)
![Icon](readme_images/icon-dark.svg#gh-dark-mode-only)

# hupload

Hupload is a minimalist file uploader for your customers.

With technical support in mind, it simplifies the process of receiving log files
or support bundles from your users.

It is a web portal and an API for direct integration into your products.

The overall concept is that share names are random tokens that can be generated
by **hupload** and are publicly accessible so users don't have to log in to 
upload content.

Note that "Copy" buttons will not work over insecure connection (http without 
SSL)

## Screenshots

**Admin page**

![Screenshot 1](readme_images/shares-dark.png#gh-dark-mode-only)
![Screenshot 1](readme_images/shares-light.png#gh-light-mode-only)

**Share page with upload box**

![Screenshot 2](readme_images/share-dark.png#gh-dark-mode-only)
![Screenshot 2](readme_images/share-light.png#gh-light-mode-only)

**Advanced share settings**

![Screenshot 2](readme_images/properties-dark.png#gh-dark-mode-only)
![Screenshot 2](readme_images/properties-light.png#gh-light-mode-only)

**Markdown preview**

![Screenshot 2](readme_images/properties-preview-dark.png#gh-dark-mode-only)
![Screenshot 2](readme_images/properties-preview-light.png#gh-light-mode-only)

## Environment variables

| Variable    | Description |
|-------------|-------------|
| CONFIG      | Path to `config.yml`    |
| HTTP_PORT   | Port to run web service |
| JWT_SECRET  | Random string used to sign sessions cookies |

## Features

- Quickly create random links and share with users,
- Easy to use drag and drop interface for uploads,
- S3 or filesystem storage,
- Configurable max share size and max file size,
- Basic share informations listed (number of items, total size),
- Add instructions in Markdown for your users,
- Automatic dark mode following OS settings,
- Multi user (all admins see all shares, but see their own listed separately first),
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
availability_days: 12
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

Currently, there is only one authentication backend `file` and you can use
either local file storage `file` or `s3` with the following configuration :

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

## Run in a container

You can quickly test **Hupload** in a container, or run it in production :

Note that the container is built to run _hupload_ as **nonroot** user (65532),
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
