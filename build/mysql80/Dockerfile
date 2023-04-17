# This Dockerfile should be used for docker official repo

# https://github.com/docker-library/official-images:
# No official images can be derived from, or depend on, non-official images
# with the following notable exceptions...
FROM oraclelinux:8-slim as base

LABEL org.opencontainers.image.authors="info@radondb.com"


RUN set -ex; \
    # shadow-utils are needed for user/group manipulation on UBI-based images
    microdnf -y update; \
    microdnf -y install glibc-langpack-en \
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
    rm -rf /usr/share/X11/xkb/symbols/id; \
    rm -rf /etc/pki/ca-trust/extracted/pem/email-ca-bundle.pem;\
    rm -rf /etc/pki/ca-trust/extracted/pem/objsign-ca-bundle.pem;\
    rm -rf /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem;\
    rm -rf /etc/pki/tls/cert.pem;\
    rm -rf /etc/unbound/icannbundle.pem;\
    rm -rf /usr/share/doc/perl-Net-SSLeay/examples/server_key.pem;\
    rm -rf /usr/share/gnupg/sks-keyservers.netCA.pem;\
    rm -rf /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.default.crt;\
    rm -rf /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.disable.crt;\
    rm -rf /etc/pki/ca-trust/extracted/pem/email-ca-bundle.pem;\
    rm -rf /etc/pki/ca-trust/extracted/pem/objsign-ca-bundle.pem;\
    rm -rf /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem;\
    rm -rf /etc/pki/tls/cert.pem;\
    rm -rf /etc/unbound/icannbundle.pem;\
    rm -rf /usr/share/doc/perl-IO-Socket-SSL;\
    rm -rf /usr/share/gnupg/sks-keyservers.netCA.pem;\
    rm -rf /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.default.crt;\
    rm -rf /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.disable.crt;\
    rm -rf /usr/share/dbus-1/session.conf;\
    rm -rf /usr/share/dbus-1/system.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.systemd1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.timedate1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.resolve1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.locale1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.portable1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.login1.conf;\
    rm -rf /usr/share/dbus-1/system.d/org.freedesktop.hostname1.conf;\
    rm -rf /usr/share/doc/perl-Net-SSLeay/examples/req.conf;\
    rm -rf /usr/share/doc/systemd/20-yama-ptrace.conf;\
    rm -rf /usr/share/doc/gnupg2/examples/gpgconf.conf;\
    rm -rf /usr/share/factory/etc/nsswitch.conf;\
    rm -rf /usr/lib/systemd/portable/profile/trusted/service.conf;\
    rm -rf /usr/lib/systemd/portable/profile/default/service.conf;\
    rm -rf /usr/lib/systemd/portable/profile/nonetwork/service.conf;\
    rm -rf /usr/lib/systemd/portable/profile/strict/service.conf;\
    rm -rf /usr/lib/systemd/system/user-.slice.d/10-defaults.conf;\
    rm -rf /usr/lib/systemd/resolv.conf;\
    rm -rf /usr/lib/sysctl.d/50-default.conf;\
    rm -rf /usr/lib/sysctl.d/50-pid-max.conf;\
    rm -rf /usr/lib/sysctl.d/01-unprivileged-bpf.conf;\
    rm -rf /usr/lib/sysctl.d/10-default-yama-scope.conf;\
    rm -rf /usr/lib/sysctl.d/50-coredump.conf;\
    rm -rf /usr/lib/environment.d/99-environment.conf;\
    rm -rf /usr/lib/sysusers.d/systemd.conf;\
    rm -rf /usr/lib/sysusers.d/basic.conf;\
    rm -rf /usr/lib/sysusers.d/dbus.conf;\
    rm -rf /usr/lib/tmpfiles.d/libselinux.conf;\
    rm -rf /usr/lib/tmpfiles.d/cryptsetup.conf;\
    rm -rf /usr/lib/tmpfiles.d/home.conf;\
    rm -rf /usr/lib/tmpfiles.d/rpm.conf;\
    rm -rf /usr/lib/tmpfiles.d/journal-nocow.conf;\
    rm -rf /usr/lib/tmpfiles.d/pam.conf;\
    rm -rf /usr/lib/tmpfiles.d/systemd.conf;\
    rm -rf /usr/lib/tmpfiles.d/legacy.conf;\
    rm -rf /usr/lib/tmpfiles.d/mysql.conf;\
    rm -rf /usr/lib/tmpfiles.d/systemd-nologin.conf;\
    rm -rf /usr/lib/tmpfiles.d/etc.conf;\
    rm -rf /usr/lib/tmpfiles.d/var.conf;\
    rm -rf /usr/lib/tmpfiles.d/portables.conf;\
    rm -rf /usr/lib/tmpfiles.d/x11.conf;\
    rm -rf /usr/lib/tmpfiles.d/dbus.conf;\
    rm -rf /usr/lib/tmpfiles.d/dnf.conf;\
    rm -rf /usr/lib/tmpfiles.d/systemd-pstore.conf;\
    rm -rf /usr/lib/tmpfiles.d/tmp.conf;\
    rm -rf /etc/libreport/events.d/collect_dnf.conf;\
    rm -rf /etc/sysctl.conf;\
    rm -rf /etc/pki/ca-trust/ca-legacy.conf;\
    rm -rf /etc/dbus-1/session.conf;\
    rm -rf /etc/dbus-1/system.conf;\
    rm -rf /etc/systemd/coredump.conf;\
    rm -rf /etc/systemd/system.conf;\
    rm -rf /etc/systemd/user.conf;\
    rm -rf /etc/systemd/resolved.conf;\
    rm -rf /etc/systemd/logind.conf;\
    rm -rf /etc/systemd/pstore.conf;\
    rm -rf /etc/systemd/journald.conf;\
    rm -rf /etc/security/chroot.conf;\
    rm -rf /etc/security/group.conf;\
    rm -rf /etc/security/time.conf;\
    rm -rf /etc/security/pwquality.conf;\
    rm -rf /etc/security/namespace.conf;\
    rm -rf /etc/security/access.conf;\
    rm -rf /etc/security/pam_env.conf;\
    rm -rf /etc/security/sepermit.conf;\
    rm -rf /etc/security/limits.conf;\
    rm -rf /etc/security/faillock.conf;\
    rm -rf /etc/sysctl.d/99-sysctl.conf;\
    rm -rf /etc/krb5.conf;\
    rm -rf /etc/xattr.conf;\
    rm -rf /etc/ld.so.conf;\
    rm -rf /etc/nsswitch.conf;\
    rm -rf /etc/host.conf;\
    rm -rf /etc/libaudit.conf;\
    rm -rf /etc/gcrypt/random.conf;\
    rm -rf /etc/selinux/semanage.conf;\
    rm -rf /etc/ld.so.conf.d/mysql-aarch64.conf;\
    rm -rf /etc/sestatus.conf;\
    rm -rf /etc/dnf/plugins/debuginfo-install.conf;\
    rm -rf /etc/dnf/plugins/copr.conf;\
    rm -rf /etc/dnf/protected.d/setup.conf;\
    rm -rf /etc/dnf/protected.d/systemd.conf;\
    rm -rf /etc/dnf/protected.d/dnf.conf;\
    rm -rf /etc/dnf/dnf.conf;\
    rm -rf /etc/openldap/ldap.conf;\
    rm -rf /etc/csh.cshrc;

FROM scratch
COPY --from=base / /


USER mysql
EXPOSE 3306 33060
CMD ["mysqld"]
