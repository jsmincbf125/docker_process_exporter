FROM python:3.5-alpine
RUN apk --update --no-cache add docker
RUN pip install docker-pid prometheus_client
COPY exporter.py /
CMD python exporter.py