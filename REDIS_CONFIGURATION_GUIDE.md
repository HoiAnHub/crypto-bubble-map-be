# Redis Configuration Guide

## Overview

The crypto-bubble-map backend service supports flexible Redis configuration using either individual parameters or URL-based connection strings. This guide explains how to configure Redis for different environments and use cases.

## Configuration Methods

### Method 1: Individual Parameters (Recommended)

Use individual environment variables for maximum flexibility and control:

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0
```

**Benefits:**
- Clear separation of connection parameters
- Easy to override individual settings
- Better for containerized environments
- Supports database selection
- More secure password handling

### Method 2: URL-based (Backward Compatibility)

Use a single Redis URL for simple setups:

```bash
REDIS_URL=redis://localhost:6379
```

**Benefits:**
- Simple single-parameter configuration
- Compatible with many Redis hosting services
- Backward compatible with existing setups

## Environment Variables Reference

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `REDIS_HOST` | Redis server hostname | `localhost` | `redis.example.com` |
| `REDIS_PORT` | Redis server port | `6379` | `6380` |
| `REDIS_PASS` | Redis password | `undefined` | `mySecurePassword` |
| `REDIS_DB` | Redis database number | `0` | `1` |
| `REDIS_URL` | Complete Redis URL | `redis://localhost:6379` | `redis://user:pass@host:port/db` |
| `REDIS_PASSWORD` | Legacy password variable | `undefined` | `password` (deprecated, use `REDIS_PASS`) |

## Configuration Priority

The service uses the following priority order:

1. **Individual Parameters**: If `REDIS_HOST` and `REDIS_PORT` are set, use individual parameters
2. **URL Fallback**: If individual parameters are not complete, fall back to `REDIS_URL`
3. **Password Priority**: `REDIS_PASS` takes precedence over `REDIS_PASSWORD`

## Environment-Specific Configurations

### Development Environment

```bash
# .env.development
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0
```

### Testing Environment

```bash
# .env.test
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=test
REDIS_DB=1  # Use different database for tests
```

### Production Environment

```bash
# .env.production
REDIS_HOST=prod-redis.example.com
REDIS_PORT=6380
REDIS_PASS=your_secure_production_password
REDIS_DB=0
```

### Docker Compose

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --requirepass dev
    volumes:
      - redis_data:/data

  app:
    build: .
    environment:
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASS: dev
      REDIS_DB: 0
    depends_on:
      - redis
```

### Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
data:
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  REDIS_DB: "0"

---
apiVersion: v1
kind: Secret
metadata:
  name: redis-secret
type: Opaque
data:
  REDIS_PASS: <base64-encoded-password>
```

## Redis Server Setup

### Local Development

#### macOS (Homebrew)
```bash
# Install Redis
brew install redis

# Start Redis with password
redis-server --requirepass dev

# Or start as service
brew services start redis
```

#### Ubuntu/Debian
```bash
# Install Redis
sudo apt update
sudo apt install redis-server

# Configure password in /etc/redis/redis.conf
sudo nano /etc/redis/redis.conf
# Uncomment and set: requirepass dev

# Restart Redis
sudo systemctl restart redis-server
```

#### Docker
```bash
# Run Redis with password
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --requirepass dev
```

### Production Setup

#### Security Considerations
1. **Use strong passwords**: Generate secure random passwords
2. **Network security**: Use private networks or VPNs
3. **SSL/TLS**: Enable Redis SSL for encrypted connections
4. **Firewall rules**: Restrict access to Redis port
5. **Regular updates**: Keep Redis version updated

#### High Availability
```bash
# Redis Sentinel configuration
REDIS_HOST=sentinel-host
REDIS_PORT=26379
REDIS_PASS=your_password
REDIS_DB=0
```

## Testing Redis Configuration

### Manual Testing
```bash
# Test connection with redis-cli
redis-cli -h localhost -p 6379 -a dev

# Test specific database
redis-cli -h localhost -p 6379 -a dev -n 0
```

### Automated Testing
```bash
# Run the Redis connection test script
node test-redis-connection.js
```

### Application Testing
```bash
# Test with the application
npm run dev

# Check logs for Redis connection status
# Look for: "Connected to Redis at localhost:6379 (DB: 0)"
```

## Troubleshooting

### Common Issues

#### Connection Refused
```
Error: connect ECONNREFUSED 127.0.0.1:6379
```
**Solutions:**
- Verify Redis server is running
- Check host and port configuration
- Verify firewall settings

#### Authentication Failed
```
Error: WRONGPASS invalid username-password pair
```
**Solutions:**
- Verify `REDIS_PASS` is correct
- Check Redis server password configuration
- Ensure password contains no special characters that need escaping

#### Database Selection Failed
```
Error: ERR DB index is out of range
```
**Solutions:**
- Verify `REDIS_DB` is within allowed range (usually 0-15)
- Check Redis server database configuration

#### SSL/TLS Issues
```
Error: unable to verify the first certificate
```
**Solutions:**
- Use `rediss://` URL for SSL connections
- Configure SSL certificates properly
- For development, consider disabling SSL verification (not recommended for production)

### Debug Commands

```bash
# Check Redis server info
redis-cli -h localhost -p 6379 -a dev INFO

# Monitor Redis commands
redis-cli -h localhost -p 6379 -a dev MONITOR

# Check memory usage
redis-cli -h localhost -p 6379 -a dev INFO memory

# List all keys (use carefully in production)
redis-cli -h localhost -p 6379 -a dev KEYS "*"
```

## Performance Optimization

### Connection Pooling
The service automatically handles connection pooling through the Redis client library.

### Memory Optimization
```bash
# Configure Redis memory policy
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### Persistence Configuration
```bash
# For job queue data, consider:
save 900 1      # Save if at least 1 key changed in 900 seconds
save 300 10     # Save if at least 10 keys changed in 300 seconds
save 60 10000   # Save if at least 10000 keys changed in 60 seconds
```

## Migration Guide

### From URL to Individual Parameters

**Before:**
```bash
REDIS_URL=redis://localhost:6379
```

**After:**
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=
REDIS_DB=0
```

### From Individual Parameters to URL

**Before:**
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=dev
REDIS_DB=0
```

**After:**
```bash
REDIS_URL=redis://:dev@localhost:6379/0
```

## Best Practices

1. **Use individual parameters** for new deployments
2. **Set passwords** even in development environments
3. **Use different databases** for different environments
4. **Monitor Redis performance** and memory usage
5. **Implement proper error handling** in application code
6. **Use SSL/TLS** in production environments
7. **Regular backups** for persistent data
8. **Monitor connection counts** and set appropriate limits

## Support

For Redis configuration issues:
1. Check the application logs for detailed error messages
2. Run `node test-redis-connection.js` to diagnose connection problems
3. Verify Redis server status and configuration
4. Check network connectivity and firewall rules
5. Review environment variable configuration
