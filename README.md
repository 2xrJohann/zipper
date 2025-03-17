WITH WRITER CHANNEL

curl --location --request POST 'http://localhost:8080/zip/with-writer-channel' \
--header 'Content-Type: application/json' \
--data-raw '{
    "paths": ["source_directory"]
}'


WITH MUTEX

curl --location --request POST 'http://localhost:8080/zip/with-mutex' \
--header 'Content-Type: application/json' \
--data-raw '{
    "paths": ["source_directory"]
}'
