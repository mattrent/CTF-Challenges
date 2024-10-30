# docker build -t test . && docker run --rm --network=host -e CHALLENGE_URL="http://localhost:8089" test

FROM alpine:3.20

RUN apk add curl

WORKDIR /app

COPY solver.sh .
RUN chmod +x solver.sh

ENTRYPOINT ["sh", "solver.sh"]
