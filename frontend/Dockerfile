# Use the official Node.js image to build the React app
FROM node:16 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy only the package.json and package-lock.json for dependency installation
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application code
COPY . .

# Build the React application
RUN npm run build

# Use the Nginx image to serve the built React app
FROM nginx:latest

# Copy the built app from the builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Expose port 80 for the web server
EXPOSE 80

# Start Nginx when the container starts
CMD ["nginx", "-g", "daemon off;"]
