FROM ubuntu:20.04

COPY webook /app/webook
WORKDIR /app

ENV WECHAT_APP_ID=001
ENV WECHAT_APP_SECRET=secret

CMD ["./webook"]