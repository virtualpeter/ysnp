FROM scratch
ARG LISTEN=:8080
ENV LISTEN=${LISTEN}
ADD ysnp /
EXPOSE ${LISTEN}
CMD ["/ysnp","-listen","${LISTEN}"]