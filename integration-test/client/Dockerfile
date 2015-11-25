FROM trsouz/ssh

ADD id_rsa /root/.ssh/id_rsa
ADD ssh_config /root/.ssh/config
ADD easyssh /usr/local/bin/easyssh
RUN chown root /root/.ssh/id_rsa /root/.ssh/config
