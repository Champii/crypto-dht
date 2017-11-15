#!/bin/bash

cd client
# npm install
npm run build
cd ..
rm -r $(pwd)/resources/app
ln -s $(pwd)/client/dist $(pwd)/resources/app
astilectron-bundler && chmod +x ./build/linux-amd64/crypto-dht
