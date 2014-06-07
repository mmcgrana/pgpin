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

chown -r vagrant:vagrant /home/vagrant/src

# Postgres
cat > /etc/apt/sources.list.d/pgdg.list <<EOF
deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main
EOF

wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt-get update

apt-get install -y postgresql-9.3

sudo -u postgres psql -U postgres -d postgres -c "alter user postgres with password 'secret';"
sudo -u postgres createdb pgpin-development
sudo -u postgres createdb pgpin-test

cat >> /home/vagrant/.profile <<EOF
export DEVELOPMENT_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-development
export TEST_DATABASE_URL=postgres://postgres:secret@127.0.0.1:5432/pgpin-test
export DATABASE_URL=\$DATABASE_URL
EOF

# Goreman
sudo -u vagrant -i go get github.com/mattn/goreman

# Config
cat >> /home/vagrant/.profile <<EOF
export API_AUTH="client:"$(openssl rand -hex 12)
export PGPIN_API_URL=http://\$API_AUTH@127.0.0.1:5000
EOF

# App dir
cd ~/src/github.com/mmcgrana/pgpin
