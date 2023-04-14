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
    rm -rf /usr/share/X11/xkb/symbols/id; \
    chmod 0600 /etc/dbus-1/session.conf;\
    chmod 0600 /etc/dbus-1/system.conf;\
    chmod 0600 /etc/dnf/dnf.conf;\
    chmod 0600 /etc/dnf/plugins/copr.conf;\
    chmod 0600 /etc/dnf/plugins/debuginfo-install.conf;\
    chmod 0600 /etc/dnf/protected.d/dnf.conf;\
    chmod 0600 /etc/dnf/protected.d/setup.conf;\
    chmod 0600 /etc/dnf/protected.d/systemd.conf;\
    chmod 0600 /etc/gcrypt/random.conf;\
    chmod 0600 /etc/host.conf;\
    chmod 0600 /etc/krb5.conf;\
    chmod 0600 /etc/ld.so.conf;\
    chmod 0600 /etc/ld.so.conf.d/mysql-aarch64.conf;\
    chmod 0600 /etc/libaudit.conf;\
    chmod 0600 /etc/libreport/events.d/collect_dnf.conf;\
    chmod 0600 /etc/nsswitch.conf;\
    chmod 0600 /etc/openldap/ldap.conf;\
    chmod 0600 /etc/pki/ca-trust/ca-legacy.conf;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/email-ca-bundle.pem;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/objsign-ca-bundle.pem;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem;\
    chmod 0600 /etc/pki/tls/cert.pem;\
    chmod 0600 /etc/security/access.conf;\
    chmod 0600 /etc/security/chroot.conf;\
    chmod 0600 /etc/security/faillock.conf;\
    chmod 0600 /etc/security/group.conf;\
    chmod 0600 /etc/security/limits.conf;\
    chmod 0600 /etc/security/namespace.conf;\
    chmod 0600 /etc/security/pam_env.conf;\
    chmod 0600 /etc/security/pwquality.conf;\
    chmod 0600 /etc/security/sepermit.conf;\
    chmod 0600 /etc/security/time.conf;\
    chmod 0600 /etc/selinux/semanage.conf;\
    chmod 0600 /etc/sestatus.conf;\
    chmod 0600 /etc/sysctl.conf;\
    chmod 0600 /etc/sysctl.d/99-sysctl.conf;\
    chmod 0600 /etc/systemd/coredump.conf;\
    chmod 0600 /etc/systemd/journald.conf;\
    chmod 0600 /etc/systemd/logind.conf;\
    chmod 0600 /etc/systemd/pstore.conf;\
    chmod 0600 /etc/systemd/resolved.conf;\
    chmod 0600 /etc/systemd/system.conf;\
    chmod 0600 /etc/systemd/user.conf;\
    chmod 0600 /etc/unbound/icannbundle.pem;\
    chmod 0600 /etc/xattr.conf;\
    chmod 0600 /usr/lib/environment.d/99-environment.conf;\
    chmod 0600 /usr/lib/sysctl.d/01-unprivileged-bpf.conf;\
    chmod 0600 /usr/lib/sysctl.d/10-default-yama-scope.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-coredump.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-default.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-pid-max.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/default/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/nonetwork/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/strict/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/trusted/service.conf;\
    chmod 0600 /usr/lib/systemd/resolv.conf;\
    chmod 0600 /usr/lib/systemd/system/user-.slice.d/10-defaults.conf;\
    chmod 0600 /usr/lib/sysusers.d/basic.conf;\
    chmod 0600 /usr/lib/sysusers.d/dbus.conf;\
    chmod 0600 /usr/lib/sysusers.d/systemd.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/cryptsetup.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/dbus.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/dnf.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/etc.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/home.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/journal-nocow.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/legacy.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/libselinux.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/mysql.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/pam.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/portables.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/rpm.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd-nologin.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd-pstore.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/tmp.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/var.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/x11.conf;\
    chmod 0600 /usr/share/dbus-1/session.conf;\
    chmod 0600 /usr/share/dbus-1/system.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.hostname1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.locale1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.login1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.portable1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.resolve1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.systemd1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.timedate1.conf;\
    chmod 0600 /usr/share/doc/gnupg2/examples/gpgconf.conf;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/client-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/client-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/proxyca.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server2-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server2-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-ecc-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-ecc-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-wildcard.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/sub-server.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/test-ca.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/test-subca.pem;\
    chmod 0600 /usr/share/doc/perl-Net-SSLeay/examples/req.conf;\
    chmod 0600 /usr/share/doc/perl-Net-SSLeay/examples/server_key.pem;\
    chmod 0600 /usr/share/doc/systemd/20-yama-ptrace.conf;\
    chmod 0600 /usr/share/factory/etc/nsswitch.conf;\
    chmod 0600 /usr/share/gnupg/sks-keyservers.netCA.pem;\
    chmod 0600 /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.default.crt;\
    chmod 0600 /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.disable.crt;\
    chmod 0600 /etc/dbus-1/session.conf;\
    chmod 0600 /etc/dbus-1/system.conf;\
    chmod 0600 /etc/dnf/dnf.conf;\
    chmod 0600 /etc/dnf/plugins/copr.conf;\
    chmod 0600 /etc/dnf/plugins/debuginfo-install.conf;\
    chmod 0600 /etc/dnf/protected.d/dnf.conf;\
    chmod 0600 /etc/dnf/protected.d/setup.conf;\
    chmod 0600 /etc/dnf/protected.d/systemd.conf;\
    chmod 0600 /etc/gcrypt/random.conf;\
    chmod 0600 /etc/host.conf;\
    chmod 0600 /etc/krb5.conf;\
    chmod 0600 /etc/ld.so.conf;\
    chmod 0600 /etc/ld.so.conf.d/mysql-aarch64.conf;\
    chmod 0600 /etc/libaudit.conf;\
    chmod 0600 /etc/libreport/events.d/collect_dnf.conf;\
    chmod 0600 /etc/nsswitch.conf;\
    chmod 0600 /etc/openldap/ldap.conf;\
    chmod 0600 /etc/pki/ca-trust/ca-legacy.conf;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/email-ca-bundle.pem;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/objsign-ca-bundle.pem;\
    chmod 0600 /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem;\
    chmod 0600 /etc/pki/tls/cert.pem;\
    chmod 0600 /etc/security/access.conf;\
    chmod 0600 /etc/security/chroot.conf;\
    chmod 0600 /etc/security/faillock.conf;\
    chmod 0600 /etc/security/group.conf;\
    chmod 0600 /etc/security/limits.conf;\
    chmod 0600 /etc/security/namespace.conf;\
    chmod 0600 /etc/security/pam_env.conf;\
    chmod 0600 /etc/security/pwquality.conf;\
    chmod 0600 /etc/security/sepermit.conf;\
    chmod 0600 /etc/security/time.conf;\
    chmod 0600 /etc/selinux/semanage.conf;\
    chmod 0600 /etc/sestatus.conf;\
    chmod 0600 /etc/sysctl.conf;\
    chmod 0600 /etc/sysctl.d/99-sysctl.conf;\
    chmod 0600 /etc/systemd/coredump.conf;\
    chmod 0600 /etc/systemd/journald.conf;\
    chmod 0600 /etc/systemd/logind.conf;\
    chmod 0600 /etc/systemd/pstore.conf;\
    chmod 0600 /etc/systemd/resolved.conf;\
    chmod 0600 /etc/systemd/system.conf;\
    chmod 0600 /etc/systemd/user.conf;\
    chmod 0600 /etc/unbound/icannbundle.pem;\
    chmod 0600 /etc/xattr.conf;\
    chmod 0600 /usr/lib/environment.d/99-environment.conf;\
    chmod 0600 /usr/lib/sysctl.d/01-unprivileged-bpf.conf;\
    chmod 0600 /usr/lib/sysctl.d/10-default-yama-scope.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-coredump.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-default.conf;\
    chmod 0600 /usr/lib/sysctl.d/50-pid-max.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/default/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/nonetwork/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/strict/service.conf;\
    chmod 0600 /usr/lib/systemd/portable/profile/trusted/service.conf;\
    chmod 0600 /usr/lib/systemd/resolv.conf;\
    chmod 0600 /usr/lib/systemd/system/user-.slice.d/10-defaults.conf;\
    chmod 0600 /usr/lib/sysusers.d/basic.conf;\
    chmod 0600 /usr/lib/sysusers.d/dbus.conf;\
    chmod 0600 /usr/lib/sysusers.d/systemd.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/cryptsetup.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/dbus.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/dnf.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/etc.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/home.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/journal-nocow.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/legacy.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/libselinux.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/mysql.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/pam.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/portables.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/rpm.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd-nologin.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/systemd-pstore.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/tmp.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/var.conf;\
    chmod 0600 /usr/lib/tmpfiles.d/x11.conf;\
    chmod 0600 /usr/share/dbus-1/session.conf;\
    chmod 0600 /usr/share/dbus-1/system.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.hostname1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.locale1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.login1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.portable1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.resolve1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.systemd1.conf;\
    chmod 0600 /usr/share/dbus-1/system.d/org.freedesktop.timedate1.conf;\
    chmod 0600 /usr/share/doc/gnupg2/examples/gpgconf.conf;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/client-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/client-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/proxyca.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server2-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server2-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-ecc-cert.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-ecc-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-key.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/server-wildcard.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/sub-server.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/test-ca.pem;\
    chmod 0600 /usr/share/doc/perl-IO-Socket-SSL/certs/test-subca.pem;\
    chmod 0600 /usr/share/doc/perl-Net-SSLeay/examples/req.conf;\
    chmod 0600 /usr/share/doc/perl-Net-SSLeay/examples/server_key.pem;\
    chmod 0600 /usr/share/doc/systemd/20-yama-ptrace.conf;\
    chmod 0600 /usr/share/factory/etc/nsswitch.conf;\
    chmod 0600 /usr/share/gnupg/sks-keyservers.netCA.pem;\
    chmod 0600 /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.default.crt;\
    chmod 0600 /usr/share/pki/ca-trust-legacy/ca-bundle.legacy.disable.crt;

FROM scratch
COPY --from=base / /


USER mysql
EXPOSE 3306 33060
CMD ["mysqld"]
