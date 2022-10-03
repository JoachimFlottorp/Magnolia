url = "https://github.com/protocolbuffers/protobuf/releases/download/v21.7/protoc-21.7-linux-x86_32.zip"
wget $url -O prot.zip
unzip prot.zip
rm prot.zip
mv bin/protoc /usr/local/bin/
mv include/google /usr/local/include/
rm -rf bin/ include/
