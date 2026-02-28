FROM mongo:latest

WORKDIR /zbc

# Copy pre-dumped dataset files into the image
COPY ./DB_gen /zbc/DB_gen

# Restore all datasets on container start
CMD mongorestore --db Crime_USENIX_REV /zbc/DB_gen/Crime_USENIX_REV/ && \
    mongorestore --db Crime_USENIX_REV_TOY /zbc/DB_gen/Crime_USENIX_REV_TOY/ && \
    mongorestore --db Enron_USENIX /zbc/DB_gen/Enron_USENIX/ && \
    mongorestore --db Wiki_USENIX /zbc/DB_gen/Wiki_USENIX/ && \
    mongod --bind_ip_all
