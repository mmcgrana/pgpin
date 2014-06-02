#!/bin/bash

set -ex

# Env
touch /home/vagrant/.profile
chown vagrant:vagrant /home/vagrant/.profile

# Go
apt-get update
apt-get install -y build-essential git-core mercurial bzr

wget --no-verbose https://storage.googleapis.com/golang/go1.2.2.src.tar.gz
tar -xz -f go1.2.2.src.tar.gz -C /opt
rm go1.2.2.src.tar.gz
cd /opt/go/src && ./make.bash --no-clean 2>&1

cat >> /home/vagrant/.profile <<EOF
export GOPATH=\$HOME
export PATH=\$HOME/bin:/opt/go/bin:\$PATH
EOF

# Postgres
cat > /etc/apt/sources.list.d/pgdg.list <<EOF
deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main
EOF

wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt-get update

apt-get install -y postgresql-9.3

sudo -u postgres psql -U postgres -d postgres -c "alter user postgres with password 'secret';"
sudo -u postgres createdb pgpin

cat >> /home/vagrant/.profile <<EOF
export DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin
EOF

# Goreman
sudo -u vagrant -i go get github.com/mattn/goreman
