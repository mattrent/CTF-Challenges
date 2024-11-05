# docker build -t test . && docker run --rm --network=host -e CHALLENGE_URL="http://localhost:8089" test

FROM alpine:3.20

RUN apk add curl --no-cache

WORKDIR /app

COPY solver.sh .
RUN chmod +x solver.sh

EXPOSE 80

ENTRYPOINT ["sh", "solver.sh"]
