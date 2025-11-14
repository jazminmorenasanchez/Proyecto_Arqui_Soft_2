FROM solr:9
# Crea el core y copia el schema
USER root
RUN mkdir -p /var/solr/data/sporthub_core/conf && \
    chown -R solr:solr /var/solr/data/sporthub_core
COPY schema.xml /var/solr/data/sporthub_core/conf/schema.xml
COPY solrconfig.xml /var/solr/data/sporthub_core/conf/solrconfig.xml
COPY core.properties /var/solr/data/sporthub_core/core.properties
RUN chown -R solr:solr /var/solr/data/sporthub_core
USER solr
