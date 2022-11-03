# data-notify-server
data-notify-server, watching tipsets from Filecoin, and pushing into message queue.

### Regenerate swagger doc
swagger doc defined in router api comment.
if edited these comments, need to regenerate swagger doc.
```shell script
swag init -g cmd/data-extraction-notify/main.go
```

### Swagger doc
swagger doc please refer to
`http://127.0.0.1:7005/data-extraction-notify/swagger/index.html`

### How to make
```
make # make to see help
```
### Run
1. Config -- data-notify-server/config
    ```
    DB of Postgresql
    Lotus0 address
    MQ of redis
    ```
2. Run
```
docker run -v /home/ec2-user/data-extraction-notify/config:/etc/data-extraction-notify/conf -p 7005:7005 -d 129862287110.dkr.ecr.us-east-2.amazonaws.com/data-infra/data-notify-server:commitId
```

### Refer
1. https://docs.google.com/document/d/1TEzwF_y4pJCv8_GfnCOZjYcdcyt7pFuX4zOP0htJ7tI/edit
2. https://docs.google.com/document/d/1LaAQCv4ZPj24TA2RdvMhqEBlf_A7d_t4BHlaktFhjLo/edit
