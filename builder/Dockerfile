FROM quay.io/quarkus/ubi-quarkus-native-image:21.0-java11

# # install native-image
# RUN gu install native-image

USER root

# # install maven
RUN mkdir -p /opt
RUN curl https://downloads.apache.org/maven/maven-3/3.8.1/binaries/apache-maven-3.8.1-bin.tar.gz | tar zx -C /opt && \
ln -s /opt/apache-maven-3.8.1/bin/mvn /usr/bin/mvn

# # install openjdk
RUN curl https://download.java.net/java/GA/jdk16.0.1/7147401fd7354114ac51ef3e1328291f/9/GPL/openjdk-16.0.1_linux-x64_bin.tar.gz | tar zx -C /opt

USER quarkus
ENV JAVA_HOME=/opt/jdk-16.0.1/
ENV PATH=$JAVA_HOME/bin:$PATH