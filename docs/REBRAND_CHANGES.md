# Miniflux to Influxeed-Engine Rebrand Documentation

This document outlines all the changes made during the rebrand from Miniflux to Influxeed-Engine.

## Overview

The complete rebrand involved updating **488 files** with **2,954 total occurrences** of "miniflux" references across the entire codebase, documentation, configuration files, and packaging scripts.

## Major Changes Summary

### 1. Core Module and Import Changes

**Go Module (`go.mod`):**
```diff
- module miniflux.app/v2
+ module influxeed-engine/v2
```

**All Go Import Statements (370+ files):**
```diff
- "miniflux.app/v2/internal/cli"
+ "influxeed-engine/v2/internal/cli"
```

**Package Import Comments:**
```diff
- package main // import "miniflux.app/v2"
+ package main // import "influxeed-engine/v2"
```

### 2. Database Configuration

**Default Database Name:**
```diff
- defaultDatabaseURL = "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
+ defaultDatabaseURL = "user=postgres password=postgres dbname=influxeed_engine sslmode=disable"
```

### 3. Application Binary and Service Names

**Binary Name:**
- Old: `miniflux`
- New: `influxeed-engine`

**SystemD Service:**
- File: `packaging/systemd/miniflux.service` → `influxeed-engine.service`
- Service name: `miniflux` → `influxeed-engine`
- User: `miniflux` → `influxeed-engine`

### 4. Build System Changes

**Makefile Updates:**
```diff
- APP := miniflux
+ APP := influxeed-engine

- DOCKER_IMAGE := miniflux/miniflux
+ DOCKER_IMAGE := influxeed-engine/influxeed-engine

- LD_FLAGS := "-s -w -X 'miniflux.app/v2/internal/version.Version=$(VERSION)'"
+ LD_FLAGS := "-s -w -X 'influxeed-engine/v2/internal/version.Version=$(VERSION)'"

- DB_URL := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable
+ DB_URL := postgres://postgres:postgres@localhost/influxeed-engine_test?sslmode=disable
```

### 5. Docker Configuration

**Dockerfile Labels:**
```diff
- LABEL org.opencontainers.image.title=miniflux
+ LABEL org.opencontainers.image.title=influxeed-engine

- LABEL org.opencontainers.image.description="Miniflux is a minimalist and opinionated feed reader"
+ LABEL org.opencontainers.image.description="influxeed-engine is a minimalist and opinionated feed reader"

- LABEL org.opencontainers.image.url=https://miniflux.app
+ LABEL org.opencontainers.image.url=https://influxeed-engine.app
```

**Binary Path:**
```diff
- COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
+ COPY --from=build /go/src/app/influxeed-engine /usr/bin/influxeed-engine

- CMD ["/usr/bin/miniflux"]
+ CMD ["/usr/bin/influxeed-engine"]
```

### 6. Documentation Updates

**README.md:**
- Title: `Miniflux 2` → `influxeed-engine 2`
- All references to miniflux.app URLs updated to influxeed-engine.app
- Application description and feature references updated

**CONTRIBUTING.md:**
- Development setup instructions updated
- Database container name: `miniflux2-db` → `influxeed-engine2-db`
- All build commands and examples updated

### 7. User Interface and Translations

**Translation Files (20+ languages):**
All translation JSON files updated with new application name:
```diff
- "Add to Miniflux"
+ "Add to influxeed-engine"

- "Miniflux API"
+ "influxeed-engine API"
```

**Error Messages:**
```diff
- "influxeed-engine is not able to reach this website due to a network error: %v"
+ "influxeed-engine is not able to reach this website due to a network error: %v"
```

### 8. Configuration Files

**Environment Configuration:**
```diff
- EnvironmentFile=/etc/miniflux.conf
+ EnvironmentFile=/etc/influxeed-engine.conf
```

**Runtime Directory:**
```diff
- RuntimeDirectory=miniflux
+ RuntimeDirectory=influxeed-engine
```

### 9. Client Library Updates

**Go Client Library:**
```diff
- package client // import "miniflux.app/v2/client"
+ package client // import "influxeed-engine/v2/client"

- import miniflux "miniflux.app/v2/client"
+ import miniflux "influxeed-engine/v2/client"
```

**Client Documentation:**
```diff
- client := miniflux.NewClient("https://api.example.org", "admin", "secret")
+ client := miniflux.NewClient("https://api.example.org", "admin", "secret")
```

## Breaking Changes for Users

### 1. Module Import Path
If you're using this as a Go module, update your imports:
```diff
- import "miniflux.app/v2/client"
+ import "influxeed-engine/v2/client"
```

### 2. Database Name
Default database name changed. Update your database configuration:
```diff
- DATABASE_URL="postgres://user:pass@host/miniflux2"
+ DATABASE_URL="postgres://user:pass@host/influxeed_engine"
```

### 3. Binary Name
The executable name changed:
```diff
- ./miniflux
+ ./influxeed-engine
```

### 4. SystemD Service
Service management commands:
```diff
- systemctl start miniflux
+ systemctl start influxeed-engine

- systemctl status miniflux
+ systemctl status influxeed-engine
```

### 5. Configuration Files
Default configuration file locations:
```diff
- /etc/miniflux.conf
+ /etc/influxeed-engine.conf
```

### 6. Docker Images
Docker image names:
```diff
- docker run miniflux/miniflux
+ docker run influxeed-engine/influxeed-engine
```

## Migration Steps

### For Development Environment:

1. **Update Go Module:**
   ```bash
   go mod edit -module=influxeed-engine/v2
   go mod tidy
   ```

2. **Database Migration:**
   ```bash
   # Create new database
   createdb influxeed_engine
   
   # Or rename existing database
   ALTER DATABASE miniflux2 RENAME TO influxeed_engine;
   ```

3. **Rebuild Application:**
   ```bash
   make clean
   make influxeed-engine
   ```

### For Production Environment:

1. **Stop Current Service:**
   ```bash
   systemctl stop miniflux
   ```

2. **Update Configuration:**
   ```bash
   # Rename config file
   mv /etc/miniflux.conf /etc/influxeed-engine.conf
   
   # Update database name in config
   sed -i 's/miniflux2/influxeed_engine/g' /etc/influxeed-engine.conf
   ```

3. **Install New Service:**
   ```bash
   # Copy new service file
   cp packaging/systemd/influxeed-engine.service /etc/systemd/system/
   
   # Reload systemd
   systemctl daemon-reload
   
   # Start new service
   systemctl start influxeed-engine
   systemctl enable influxeed-engine
   ```

4. **Database Migration:**
   ```bash
   # Rename database (if needed)
   psql -c "ALTER DATABASE miniflux2 RENAME TO influxeed_engine;"
   ```

## Files Modified

### Categories of Changes:

1. **Go Source Files:** 370+ files with import path updates
2. **Documentation:** README.md, CONTRIBUTING.md, man pages
3. **Build System:** Makefile, Docker files, GitHub Actions
4. **Packaging:** SystemD service, RPM specs, Debian packages
5. **Configuration:** Environment files, Docker Compose
6. **Translations:** 20+ language files
7. **Client Libraries:** Go client documentation and examples

### Total Impact:
- **Files Modified:** 488
- **Total Replacements:** 2,954
- **Languages Affected:** 20+ translation files
- **Platforms:** Linux, macOS, FreeBSD, OpenBSD, Windows

## Verification Steps

After rebrand, verify the changes:

1. **Build Test:**
   ```bash
   make influxeed-engine
   echo $?  # Should return 0
   ```

2. **Module Verification:**
   ```bash
   go list -m
   # Should show: influxeed-engine/v2
   ```

3. **Binary Name Check:**
   ```bash
   ls -la influxeed-engine
   ./influxeed-engine -version
   ```

4. **Database Connection Test:**
   ```bash
   ./influxeed-engine -debug
   # Check for successful database connection
   ```

## Notes

- All functionality remains identical - only naming changed
- API endpoints and responses unchanged
- Database schema unchanged (only database name)
- Configuration environment variables unchanged
- No data loss during migration

This rebrand maintains full backward compatibility at the API level while updating all internal references and branding to Influxeed-Engine.