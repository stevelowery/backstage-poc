export NAME=${{ values.component_id }}
export NAMESPACE=${{ values.namespace }}

# Generate the CA cert and private key
openssl req -nodes -new -days +3650 -x509 -keyout certs/ca.key -out certs/ca.crt -subj "/CN=Admission Controller Webhook CA"

# Generate the private key for the webhook server
openssl genrsa -out certs/tls.key 2048

# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key certs/tls.key -subj "/CN=${NAME}.${NAMESPACE}.svc" -addext "subjectAltName = DNS:${NAME}.${NAMESPACE}.svc" \
 | openssl x509 -req -days +3650 -extfile <(printf "subjectAltName = DNS:${NAME}.${NAMESPACE}.svc") -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -in certs/tls.csr -out certs/tls.crt