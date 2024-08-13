FROM mongo:7.0.12

RUN mkdir /config

WORKDIR /config

COPY wait-for-it.sh .
COPY setup.js .
COPY setup.sh .

RUN chmod +x /config/wait-for-it.sh
RUN chmod +x /config/setup.sh

CMD ["bash", "-c", "/config/wait-for-it.sh crm-mongo:27017 -- /config/setup.sh"]
