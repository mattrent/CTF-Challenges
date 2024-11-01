# docker build -t test . && docker run --rm -v ./../handout/challenge:/app/challenge -v ./../challenge.yml:/app/challenge.yml test

FROM python:3.13-slim-bookworm

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY solver.py .

ENTRYPOINT ["python", "solver.py"]
