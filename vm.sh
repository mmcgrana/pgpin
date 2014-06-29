#!/bin/bash

set -exo pipefail

# Env
touch /home/vagrant/.profile
chown vagrant:vagrant /home/vagrant/.profile

# Synced folder ownership
chown -R vagrant:vagrant /home/vagrant/src

# Go
apt-get update
apt-get install -y --no-install-recommends build-essential curl git-core mercurial bzr

mkdir -p /opt/go
curl -s https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz | tar xzf - -C /opt/go --strip-components=1

cat >> /home/vagrant/.profile <<EOF
export GOROOT=/opt/go
export GOPATH=\$HOME
export PATH=\$HOME/bin:/opt/go/bin:\$PATH
EOF

# Godep and Goreman
sudo -u vagrant -i go get github.com/tools/godep
sudo -u vagrant -i go get github.com/mattn/goreman

# Postgres
cat > /etc/apt/sources.list.d/pgdg.list <<EOF
deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main
EOF

curl -s https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt-get update

apt-get install -y --no-install-recommends postgresql-9.3

sudo -u postgres psql -U postgres -d postgres -c "alter user postgres with password 'secret';"
sudo -u postgres createdb pgpin-development
sudo -u postgres createdb pgpin-test

cat >> /home/vagrant/.profile <<EOF
export DEVELOPMENT_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-development
export TEST_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-test
export DATABASE_URL=\$DEVELOPMENT_DATABASE_URL
EOF

# Redis
apt-get install -y --no-install-recommends redis-server

cat >> /home/vagrant/.profile <<EOF
export DEVELOPMENT_REDIS_URL=redis://user:@127.0.0.1:6379/1
export TEST_REDIS_URL=redis://user:@127.0.0.1:6379/2
export REDIS_URL=\$DEVELOPMENT_REDIS_URL
EOF

# Other app config
cat >> /home/vagrant/.profile <<EOF
export FERNET_KEYS=$(openssl rand -base64 32)
export PGPIN_URL=http://127.0.0.1:5000
EOF

# App directory
cat >> /home/vagrant/.profile <<EOF
cd ~/src/github.com/mmcgrana/pgpin
EOF
