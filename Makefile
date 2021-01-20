TARGET := helmMgr
SRCDIR := $(PWD)
CC := go

REPO := registry.cn-shanghai.aliyuncs.com/digk8s/
TAG := v1.0.1
FLAGS := -ldflags "-s -w"

all: helmmgr-bin

helmmgr-image:
	docker build -t $(REPO)helmmgr:$(TAG) -f $(SRCDIR)/cmd/helmmgr/Dockerfile .
	docker push $(REPO)helmmgr:$(TAG)
helmmgr-strip:
	$(CC) build $(FLAGS) -o $(SRCDIR)/bin/helmmgr $(SRCDIR)/cmd/helmmgr

helmmgr-bin:
	$(CC) build -mod=vendor -o $(SRCDIR)/bin/helmmgr $(SRCDIR)/cmd/helmmgr

.PHONY:clean
clean:
	rm -rf $(SRCDIR)/bin

