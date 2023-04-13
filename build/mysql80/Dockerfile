# This Dockerfile should be used for docker official repo

# https://github.com/docker-library/official-images:
# No official images can be derived from, or depend on, non-official images
# with the following notable exceptions...
FROM oraclelinux:8-slim

LABEL org.opencontainers.image.authors="info@radondb.com"


RUN set -ex; \
    # shadow-utils are needed for user/group manipulation on UBI-based images
    microdnf -y update; \
    microdnf -y install glibc-langpack-en \
    nss_wrapper \
    epel-release \
    findutils \
    shadow-utils; \
    microdnf clean all; \
    groupadd -g 1001 mysql; \
    useradd -u 1001 -r -g 1001 -s /sbin/nologin \
    -c "Default Application User" mysql ;\
    { \
    echo '[mysql8.0-server-minimal]'; \
    echo 'name=MySQL 8.0 Server Minimal'; \
    echo 'enabled=1'; \
    echo 'baseurl=http://139.198.40.93:801' ;\
    echo 'gpgcheck=0'; \
    echo 'module_hotfixes=true' ;\
    } | tee /etc/yum.repos.d/mysql-community-minimal.repo

# check repository package signature in secure way
ENV PS_VERSION 8.0.25-15.1
ENV OS_VER el8
ENV FULL_PERCONA_VERSION "$PS_VERSION.$OS_VER"

RUN set -ex; \
    rpm -e --nodeps tzdata; \
    microdnf -y install \
    hostname \
    tzdata \
    jemalloc \
    which \
    cracklib-dicts \
    tar \
    policycoreutils; \
    \
    microdnf -y install \
    percona-server-server-${FULL_PERCONA_VERSION}; \
    microdnf clean all; \
    rm -rf /var/cache/dnf /var/cache/yum /var/lib/mysql

# purge and re-create /var/lib/mysql with appropriate ownership
RUN set -ex; \
    /usr/bin/install -m 0775 -o mysql -g root -d /var/lib/mysql /var/run/mysqld /docker-entrypoint-initdb.d; \
    # comment out a few problematic configuration values
    find /etc/my.cnf /etc/my.cnf.d -name '*.cnf' -print0 \
    | xargs -0 grep -lZE '^(bind-address|log|user)' \
    | xargs -rt -0 sed -Ei 's/^(bind-address|log|user)/#&/'; \
    # don't reverse lookup hostnames, they are usually another container
    echo '!includedir /etc/my.cnf.d' >> /etc/my.cnf; \
    printf '[mysqld]\nskip-host-cache\nskip-name-resolve\n' > /etc/my.cnf.d/docker.cnf; \
    # TokuDB modifications
    /usr/bin/install -m 0664 -o mysql -g root /dev/null /etc/sysconfig/mysql; \
    echo "LD_PRELOAD=/usr/lib64/libjemalloc.so.1" >> /etc/sysconfig/mysql; \
    echo "THP_SETTING=never" >> /etc/sysconfig/mysql; \
    # allow to change config files
    chown -R mysql:root /etc/my.cnf /etc/my.cnf.d; \
    chmod -R ug+rwX /etc/my.cnf /etc/my.cnf.d
RUN microdnf  --enablerepo='*'  update -y;mysql --version \
    microdnf clean all;
VOLUME ["/var/lib/mysql", "/var/log/mysql"]

COPY build/mysql80/ps-entry.sh /docker-entrypoint.sh
RUN chmod +rx /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]
RUN rm -rf /usr/bin/id;\
    rm -rf /usr/bin/make;\
    rm -rf /usr/bin/perl;\
    rm -rf /usr/share/cmake/Modules/Compiler/XL-Fortran/cpp;\
    rm -rf /usr/share/doc/make;\
    rm -rf /usr/share/gdb;\
    rm -rf /usr/share/licenses/make;\
    rm -rf /usr/share/locale/id;\
    rm -rf /usr/share/perl5/unicore/lib/Perl;\
    rm -rf /usr/share/X11/xkb/symbols/id;



USER mysql
EXPOSE 3306 33060
CMD ["mysqld"]
