mkfile_path_build := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir_build := $(dir $(mkfile_path_build))

.SILENT:

clean:
	rm -Rf ${mkfile_dir_build}/alpine-base
	rm -Rf ${mkfile_dir_build}/alpine-final

.PHONY: alpine
alpine:
	if [ ! -d ${mkfile_dir_build}/alpine-base ]; then \
		vorteil projects convert-container alpine ${mkfile_dir_build}/alpine-base; \
	fi


# make DB=true prep-alpine to install postgresql
.PHONY: prep-alpine
prep-alpine: alpine
prep-alpine: docker-ui
		cp ${mkfile_dir_build}/builder.sh ${mkfile_dir_build}/alpine-base
		cp ${mkfile_dir_build}/builder-pg.sh ${mkfile_dir_build}/alpine-base
		if [ ! -d ${mkfile_dir_build}/alpine-final ]; then \
			vorteil run --program[1].env="PATH=/usr/bin:/bin:/sbin" --program[1].args="/builder-pg.sh $(DB)" --program[1].privilege="superuser" --program[1].bootstrap="WAIT_FILE /done" \
				--system.user="direktiv" --program[0].args="/builder.sh $(DB)" --vm.disk-size="+1024 MiB" ${mkfile_dir_build}/alpine-base --record=${mkfile_dir_build}/alpine-final; \
		fi
		cp ${mkfile_dir_build}/conf.toml ${mkfile_dir_build}/alpine-final/etc
		cp ${mkfile_dir_build}/../direktiv ${mkfile_dir_build}/alpine-final/bin
		cp ${mkfile_dir_build}/docker/ui/direktiv-ui ${mkfile_dir_build}/alpine-final/bin

		cp -f ${mkfile_dir_build}/default.vcfg ${mkfile_dir_build}/alpine-final
		cp -f ${mkfile_dir_build}/ui.vcfg ${mkfile_dir_build}/alpine-final
		cp -f ${mkfile_dir_build}/vorteilproject ${mkfile_dir_build}/alpine-final/.vorteilproject
		rm -f ${mkfile_dir_build}/alpine-final/var/lib/postgresql/data/postmaster.pid

# Run with DB:
# 'make DB_HOST=127.0.0.1 DB_USER=direktiv DB_PWD=direktiv DB=true run-alpine' for connecting to host database running on vorteil
# 'make DB_HOST=192.168.1.10 DB_USER=sisatech DB_PWD=sisatech run-alpine' for connecting to host database runningn on vorteil
.PHONY: run-alpine
run-alpine: prep-alpine
run-alpine:
		rm -f ${mkfile_dir_build}/alpine-final/var/lib/postgresql/data/postmaster.pid
		vorteil run --program[2].env="DIREKTIV_SECRETS_DB=host=$(DB_HOST) port=5432 user=$(DB_USER) dbname=postgres password=$(DB_PWD) sslmode=disable" --program[2].env="DIREKTIV_DB=host=$(DB_HOST) port=5432 user=$(DB_USER) dbname=postgres password=$(DB_PWD) sslmode=disable"	${mkfile_dir_build}/alpine-final
