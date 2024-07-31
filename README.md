# hupload

Hupload is a minimalist file uploader for your customers.

With technical support in mind, it simplifies the process of receiving log files
or support bundles from your users.

It is a web portal and an API for direct integration into your products.

## Environment variables

| Variable    | Description |
|-------------|-------------|
| CONFIG      | Path to `config.yml`    |
| HTTP_PORT   | Port to run web service |

## Configuration

By default, Hupload uses `data/` directory and `admin` user with a randomly
generated password that is displayed in the logs during startup.

If a configuration file is provided in `CONFIG` environment variable, it will
be used to configure storage backend and authentication.

Sample configuration file :

```
auth:
  type: file
  options:
    path: config/users.yml
backend:
  type: file
  options:
    path: data
```

Currently, there is only one authentication backend `file` and one storage
backend `file`

For auehtentication, users are defined in a yaml file :

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
