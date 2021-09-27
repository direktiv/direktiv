FROM ubuntu:20.04

RUN apt-get update && apt-get install -y python3 wget openssh-client openssh-server curl

RUN wget https://downloads.yugabyte.com/yugabyte-2.9.0.0-linux.tar.gz

RUN tar -xvzf yugabyte-2.9.0.0-linux.tar.gz && /yugabyte-2.9.0.0/bin/post_install.sh

CMD ["bash"]
