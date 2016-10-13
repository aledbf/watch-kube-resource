all: push

# 0.0 shouldn't clobber any release builds
TAG = 0.1
BINARY_NAME = watch-resource-cmdrunner
PREFIX = aledbf/$(BINARY_NAME)

binary: clean
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o $(BINARY_NAME)

container: binary
	docker build -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	rm -f $(BINARY_NAME)
