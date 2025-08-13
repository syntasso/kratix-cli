FROM python:3.12-slim

RUN apt-get update && apt-get install -y git

RUN python -m pip install git+https://github.com/syntasso/kratix-python.git

COPY scripts/pipeline.py /usr/bin/pipeline.py

CMD [ "sh", "-c", "pipeline.py" ]

ENTRYPOINT []