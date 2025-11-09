# Multi-stage build for React frontend

# Stage 1: Build the React application
FROM node:20-alpine AS build

WORKDIR /app

# Accept build arguments for environment variables
ARG VITE_GITHUB_CLIENT_ID
ARG VITE_API_URL=http://localhost:3000

# Set environment variables for Vite build
ENV VITE_GITHUB_CLIENT_ID=$VITE_GITHUB_CLIENT_ID
ENV VITE_API_URL=$VITE_API_URL

# Copy package files first (for better layer caching)
COPY package.json package-lock.json* ./

# Install dependencies (cached unless package files change)
RUN npm ci

# Copy source code (this layer changes frequently)
COPY . .

# Build the application with timestamp to prevent caching
RUN npm run build && echo "Build completed at $(date)" > /app/dist/.build-timestamp

# Stage 2: Serve with nginx
FROM nginx:alpine

# Install curl for healthchecks
RUN apk add --no-cache curl

# Copy built assets from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy nginx configuration for SPA routing
COPY <<EOF /etc/nginx/conf.d/default.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # SPA routing - try files first, then fallback to index.html
    location / {
        try_files \$uri \$uri/ /index.html;
        
        # NEVER cache HTML files (they reference hashed assets)
        add_header Cache-Control "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    # Cache hashed static assets aggressively (filenames change = cache busted)
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
