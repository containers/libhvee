# e2e

TBC

## Container

Build container image

```bash
make build-oci-e2e
```

Sample for running the container with e2e on remote target

```bash
podman run --rm -it --name libhvee-e2e \
    -e TARGET_HOST=$(cat host) \
    -e TARGET_HOST_USERNAME=$(cat username) \
    -e TARGET_HOST_KEY_PATH=/data/id_rsa \
    -e TARGET_FOLDER=libhvee-e2e \
    -e TARGET_RESULTS=libhvee-e2e.xml \
    -e OUTPUT_FOLDER=/data \
    -e DEBUG=true \
    -v $PWD:/data:z \
    quay.io/rhqp/libhvee-e2e:v0.0.1 \
        libhvee-e2e/run.ps1 \
            -targetFolder libhvee-e2e \
            -junitResultsFilename libhvee-e2e.xml 
```