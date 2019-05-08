#!/bin/bash

# Change to the directory with our code that we plan to work from
cd "$GOPATH/src/lenslocked.com"

echo "==== Releasing lenslocked.com ===="
echo "  Deleting the local binary if it exists (so it isn't uploaded)..."
rm lenslocked.com
echo "  Done!"

echo "  Deleting existing code..."
ssh root@screencast.lenslocked.com "rm -rf /root/go/src/lenslocked.com"
echo "  Code deleted successfully!"

echo "  Uploading code..."
rsync -avr --exclude '.git/*' --exclude 'tmp/*' --exclude 'images/*' ./ root@screencast.lenslocked.com:/root/go/src/lenslocked.com/
echo "  Code uploaded successfully!"

echo "  Go getting deps..."
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get golang.org/x/crypto/bcrypt"
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get github.com/gorilla/mux"
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get github.com/gorilla/schema"
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get github.com/lib/pq"
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get github.com/jinzhu/gorm"
ssh root@screencast.lenslocked.com "export GOPATH=/root/go; /usr/local/go/bin/go get github.com/gorilla/csrf"

echo "  Building the code on remote server..."
ssh root@screencast.lenslocked.com 'export GOPATH=/root/go; cd /root/app; /usr/local/go/bin/go build -o ./server $GOPATH/src/lenslocked.com/*.go'
echo "  Code built successfully!"

echo "  Moving assets..."
ssh root@screencast.lenslocked.com "cd /root/app; cp -R /root/go/src/lenslocked.com/assets ."
echo "  Assets moved successfully!"

echo "  Moving views..."
ssh root@screencast.lenslocked.com "cd /root/app; cp -R /root/go/src/lenslocked.com/views ."
echo "  Views moved successfully!"

echo "  Moving Caddyfile..."
ssh root@screencast.lenslocked.com "cd /root/app; cp /root/go/src/lenslocked.com/Caddyfile ."
echo "  Views moved successfully!"

echo "  Restarting the server..."
ssh root@screencast.lenslocked.com "sudo service lenslocked.com restart"
echo "  Server restarted successfully!"

echo "  Restarting Caddy server..."
ssh root@screencast.lenslocked.com "sudo service caddy restart"
echo "  Caddy restarted successfully!"

echo "==== Done releasing lenslocked.com ===="
