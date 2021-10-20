FROM golang:1.16.9-alpine3.14 as builder


# Update apk ======================================================================================
RUN apk --update upgrade

# Intall essential packages =======================================================================
RUN apk add bash wget git

# Install build tools =============================================================================
RUN apk add build-base gcc

# Install Postgres client =========================================================================
RUN apk add postgresql-client

# Install line-ending conversion tool =============================================================
# This is only used during the duration of this Dockerfile and will be removed in the end.
RUN apk add dos2unix


# Update CA certificates
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Copy entry/support files ========================================================================
COPY docker/entry.sh /entry.sh
RUN dos2unix --quiet /entry.sh
RUN chmod +x /entry.sh

# Cleanup =========================================================================================
RUN apk del dos2unix
RUN rm -rf /var/cache/apk/*

# Setup entry point ===============================================================================
ENTRYPOINT ["/entry.sh"]
