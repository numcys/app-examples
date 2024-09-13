# Use an official Node.js runtime as a parent image
FROM node:18

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /usr/src/app
COPY package*.json ./
COPY upload-timestamp.js .

# Install any needed packages specified in package.json
RUN npm install

# Make port 8080 available to the world outside this container (optional, if you need networking)
EXPOSE 8080

# Run the script
CMD ["node", "upload-timestamp.js"]
