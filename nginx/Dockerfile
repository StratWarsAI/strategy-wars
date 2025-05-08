FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Create directories for SSL certificates
RUN mkdir -p /etc/nginx/ssl

# Copy custom nginx config
COPY nginx-cloudflare.conf /etc/nginx/nginx.conf

# Copy Cloudflare Origin certificates
COPY origin.pem /etc/nginx/ssl/origin.pem
COPY private.key /etc/nginx/ssl/private.key

# Create directory for nginx pid file
RUN mkdir -p /var/run/nginx

# Expose ports
EXPOSE 80
EXPOSE 443

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]