mkdir -p certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout certs/www.confusion.com.key -out certs/www.confusion.com.crt \
    -subj "/C=CA/ST=Alberta/L=Calgary/O=Brian Woo/OU=Development/CN=www.confusion.com/emailAddress=dummy@mail.com"
