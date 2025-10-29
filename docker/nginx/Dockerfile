FROM nginx:1.25-alpine

# Use Alpine-based nginx to avoid Debian CVE surface area
USER root
# Refresh OS packages to pick up security fixes from Alpine repos
# and remove any accidental snakeoil certs if present
RUN apk update && apk upgrade --no-cache \
	&& rm -rf /var/cache/apk/* \
	&& rm -f /etc/ssl/private/ssl-cert-snakeoil.key || true

USER nginx

CMD ["nginx", "-g", "daemon off;"]
