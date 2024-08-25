#EJ2
FROM python:3.9.7-slim
COPY server /server
ENV PYTHONPATH=/server
RUN python -m unittest /server/tests/test_common.py
WORKDIR /server
ENTRYPOINT ["/bin/sh"]
