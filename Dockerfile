FROM drupal:10-apache

COPY script.sh /script.sh

CMD ["/bin/bash", "-c", "/script.sh ","&&"," apache2-foreground"]
