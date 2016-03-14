FROM concourse/busyboxplus:hg

ENV LANG C

RUN mkdir -p /opt/resource
ADD hgresource/hgresource /opt/resource
RUN chmod +x /opt/resource/*
RUN ln -s /opt/resource/hgresource /opt/resource/in; ln -s /opt/resource/hgresource /opt/resource/out; ln -s /opt/resource/hgresource /opt/resource/check

ADD test/ /opt/resource-tests/

RUN /opt/resource-tests/all.sh && \
  rm -rf /tmp/*
