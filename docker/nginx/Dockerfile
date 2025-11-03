FROM nginx:1.25-alpine

# Use Alpine-based nginx to avoid Debian CVE surface area
USER root
# Refresh OS packages to pick up security fixes from Alpine repos
# and remove any accidental snakeoil certs if present
RUN apk update && apk upgrade --no-cache \
	&& rm -rf /var/cache/apk/* \
	&& rm -f /etc/ssl/private/ssl-cert-snakeoil.key || true

# Create all nginx required directories with proper permissions before switching user
RUN mkdir -p /var/cache/nginx/client_temp \
	&& mkdir -p /var/cache/nginx/proxy_temp \
	&& mkdir -p /var/cache/nginx/fastcgi_temp \
	&& mkdir -p /var/cache/nginx/uwsgi_temp \
	&& mkdir -p /var/cache/nginx/scgi_temp \
	&& mkdir -p /var/run/nginx \
	&& chown -R nginx:nginx /var/cache/nginx /var/run/nginx /var/log/nginx \
	&& chmod -R 755 /var/cache/nginx /var/run/nginx /var/log/nginx

# Copy custom nginx.conf with correct pid path
COPY nginx.conf /etc/nginx/nginx.conf
RUN chown nginx:nginx /etc/nginx/nginx.conf && chmod 644 /etc/nginx/nginx.conf

# Copy routing configuration
COPY conf.d/default.conf /etc/nginx/conf.d/default.conf
RUN chown nginx:nginx /etc/nginx/conf.d/default.conf && chmod 644 /etc/nginx/conf.d/default.conf

# Ensure static dir exists. In development we mount a host directory
# (./docker/nginx/static) into /usr/share/nginx/html/static via docker-compose.
# Avoid copying app static during image build to keep build context small.
RUN mkdir -p /usr/share/nginx/html/static \
	&& chown -R nginx:nginx /usr/share/nginx/html/static \
	&& chmod -R 755 /usr/share/nginx/html/static

USER nginx

CMD ["nginx", "-g", "daemon off;"]
