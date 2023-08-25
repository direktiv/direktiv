# Building Direktiv Frontend

1. Building UI Files with Docker

There is a Makefile target to build the UI files called `build-react`. This builds the React components and files and stores them in the folder `dist`. 


2. Building 

The Makefile target `make local` can be used to create a local image. The following variables change the Docker repository, image name or tag.

- DOCKER_REPO
- DOCKER_IMAGE
- DOCKER_TAG

The following command runs the local image and the configuration can be overwritten with `DIREKTIV_` environment variables.

```
docker run --network=host -p 2304:2304 -e DIREKTIV_SERVER_BACKEND=http://MY-API-ENDPOINT localhost:5000/direktiv-frontend
```

Important Environment Variables:

DIREKTIV_LOG_DEBUG: The value `debug` enables debug logging.
DIREKTIV_LOG_JSON: Disable/enable JSON logging. Default `false`
DIREKTIV_SERVER_BACKEND: API backend for the UI 

3. Multi-Arch Building

The Makefile target `cross` is for multi-arch builds. Local repositories are not supported. Therefore the variable `DOCKER_REPO` is required.


```
make cross DOCKER_REPO=direktiv
```


