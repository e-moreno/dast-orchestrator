FROM zaproxy/zap-stable:2.16.1 as test
CMD [ "echo", "test target stage" ]

FROM zaproxy/zap-stable:2.16.1 as app

WORKDIR /zap
COPY ./conf/conf /zap

CMD ["bash", "-c", "/zap/zap.sh -daemon -port 8090 -host 0.0.0.0 -configfile conf"]