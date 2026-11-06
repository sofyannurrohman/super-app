# Production image with Nginx serving pre-built static files (0 RAM compile overhead)
FROM nginx:stable-alpine

WORKDIR /usr/share/nginx/html

# Clean default Nginx files and copy pre-built dist
RUN rm -rf ./*
COPY dist/ .

# Copy custom nginx config to handle SPA routing and API proxy
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
