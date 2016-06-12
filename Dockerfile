FROM concourse/buildroot:hg

RUN mkdir -p /opt/resource
ADD hgresource/hgresource /opt/resource
ADD assets/askpass.sh /opt/resource
RUN chmod +x /opt/resource/*
RUN ln -s /opt/resource/hgresource /opt/resource/in; ln -s /opt/resource/hgresource /opt/resource/out; ln -s /opt/resource/hgresource /opt/resource/check
