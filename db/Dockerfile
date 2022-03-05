FROM postgres:14-alpine
LABEL authors="Sergey Molchanov <sempost@gmail.com>"

COPY docker-entrypoint-initdb.d/ /docker-entrypoint-initdb.d/

HEALTHCHECK --interval=15s --timeout=3s CMD pg_isready -U user || exit 1