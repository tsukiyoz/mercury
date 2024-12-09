FROM ubuntu:20.04

COPY mercury /app/mercury
WORKDIR /app

ENV WECHAT_APP_ID=001
ENV WECHAT_APP_SECRET=secret

CMD ["./mercury"]