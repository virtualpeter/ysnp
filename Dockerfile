FROM scratch

#with these defaults it will redirect to the host and path from the request, just strips port and query flags
#and sets protocol to https

ENV PORT=8080
ENV STATUS=301
ENV TARGET_HOST=""
ENV TARGET_PORT=""
ENV TARGET_PROTO="https"
ENV TARGET_PATH=""
ENV BLOCKQUERY="false"
ENV LOG="json,info"

ADD ysnp /

EXPOSE $PORT

CMD [ "/ysnp","-listen",":$PORT",\
      "-target_proto","$TARGET_PROTO",\
      "-target_host","$TARGET_HOST",\
      "-target_port","$TARGET_PORT",\
      "-target_path","$TARGET_PATH",\
      "-blockquery","$BLOCKQUERY",\
      "-log","$LOG",\
      "-status","$STATUS"\
    ]