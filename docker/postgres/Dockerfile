FROM postgres:15-alpine

# Use Alpine-based Postgres to reduce Debian CVE surface; remove any accidental private keys
USER root
RUN rm -f /etc/ssl/private/ssl-cert-snakeoil.key || true

USER postgres

CMD ["docker-entrypoint.sh", "postgres"]
