
APP_NAME   = huecontrol
ENV       ?= dev
DBNAME    ?= $(APP_NAME)_$(ENV)
DBUSER    ?= $(APP_NAME)_$(ENV)
DBPASSWD  ?= $(APP_NAME)_$(ENV)_secret

PSQL_CMD = cd sql && psql -v ON_ERROR_STOP=1
GO_SRC   = $(shell find . -type f -name '*.go')
SQL_SRC  = $(shell find sql -type f)

default: build

dev: build views
	./$(APP_NAME) -bind :3000

build: $(APP_NAME)

$(APP_NAME): $(GO_SRC)
	godep go build .

devdb:
	$(MAKE) recreatedb
	$(MAKE) schema
	$(MAKE) devdata

.PHONY: schema
schema: $(SQL_SRC)
	$(PSQL_CMD) -d $(DBNAME) -U $(DBUSER) -f ./schema.sql

.PHONY: schema
views: $(SQL_SRC)
	$(PSQL_CMD) -d $(DBNAME) -U $(DBUSER) -f ./views.sql

devdata: $(SQL_SRC)
	$(PSQL_CMD) -d $(DBNAME) -U $(DBUSER) -f ./dev.sql

recreatedb:
	$(MAKE) dropdb || true
	$(MAKE) createdb

dropdb: clearcache
	dropdb $(DBNAME)
	dropuser $(DBUSER)

clearcache:
	redis-cli flushall

createdb:
	$(PSQL_CMD) --set=dbuser=$(DBUSER) --set=dbpasswd=$(DBPASSWD) --set=dbname=$(DBNAME) -f ./db.sql

clean:
	rm -f $(APP_NAME)
