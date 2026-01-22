# Dockerfile for postgres-pep-sim
# Simulates hospital PEP database (Tasy/MV schema) for development and demo

FROM postgres:15-alpine

# Copy initialization scripts
COPY init-schema.sql /docker-entrypoint-initdb.d/01-init-schema.sql
COPY seed-data.sql /docker-entrypoint-initdb.d/02-seed-data.sql
COPY auto-insert.sh /docker-entrypoint-initdb.d/03-auto-insert.sh

# Make auto-insert script executable
RUN chmod +x /docker-entrypoint-initdb.d/03-auto-insert.sh

# Expose port 5432 (will be mapped to 5433 externally)
EXPOSE 5432

# Labels
LABEL org.opencontainers.image.source="https://github.com/sidot/sidot"
LABEL org.opencontainers.image.description="Simulated PEP Database (Tasy schema) for SIDOT"
LABEL org.opencontainers.image.version="1.0.0"
