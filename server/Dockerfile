FROM python:3.9.7-slim

COPY server /server

ENV PYTHONPATH=/server

RUN python -m unittest discover -s /server/tests -p "*.py"

WORKDIR /server

ENTRYPOINT ["python", "main.py"]
