# Base image
FROM drupal:10-apache

# Install PostgreSQL client
RUN apt-get update && apt-get install -y postgresql-client

# Copy Drupal configuration
# COPY sites /var/www/html/sites

# Set environment variables
ENV POSTGRES_HOST postgres
ENV POSTGRES_DB postgres
ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD example

# Expose the Drupal port
EXPOSE 80

# Command to run the web server
CMD ["apache2", "start"]
