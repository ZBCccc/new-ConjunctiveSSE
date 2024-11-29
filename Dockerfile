# 使用官方 MongoDB 镜像作为基础镜像
FROM mongo:latest

# 设置工作目录
WORKDIR /zbc

# 将本地数据库文件夹复制到容器中
COPY ./DB_gen /zbc

# 设置容器启动时恢复数据库的命令
CMD mongorestore --db Crime_USENIX_REV /zbc/DB_gen/Crime_USENIX_REV/ && \
    mongorestore --db Crime_USENIX_REV_TOY /zbc/DB_gen/Crime_USENIX_REV_TOY/ && \
    mongorestore --db Enron_USENIX /zbc/DB_gen/Enron_USENIX/ && \
    mongorestore --db Wiki_USENIX /zbc/DB_gen/Wiki_USENIX/
