import { createClient, RedisClientType } from 'redis';
import { config } from '@/config/config';
import { logger } from '@/utils/logger';
import { CacheEntry } from '@/types';

export class RedisService {
  private client: RedisClientType | null = null;
  private isConnected = false;

  constructor() {
    // Initialize client but don't connect yet
  }

  public async connect(): Promise<void> {
    if (this.isConnected && this.client) {
      return;
    }

    try {
      // Use individual parameters if available, fallback to URL
      const redisConfig: any = {};

      if (config.redis.host && config.redis.port) {
        // Use individual parameters
        redisConfig.socket = {
          host: config.redis.host,
          port: config.redis.port,
          reconnectStrategy: (retries: number) => {
            if (retries > 10) {
              logger.error('Redis reconnection failed after 10 attempts');
              return false;
            }
            return Math.min(retries * 100, 3000);
          },
        };

        if (config.redis.password) {
          redisConfig.password = config.redis.password;
        }

        if (config.redis.database !== undefined && config.redis.database !== 0) {
          redisConfig.database = config.redis.database;
        }

        logger.info(`Connecting to Redis at ${config.redis.host}:${config.redis.port} (DB: ${config.redis.database})`);
      } else {
        // Fallback to URL-based connection for backward compatibility
        redisConfig.url = config.redis.url;
        if (config.redis.password) {
          redisConfig.password = config.redis.password;
        }
        redisConfig.socket = {
          reconnectStrategy: (retries: number) => {
            if (retries > 10) {
              logger.error('Redis reconnection failed after 10 attempts');
              return false;
            }
            return Math.min(retries * 100, 3000);
          },
        };

        logger.info(`Connecting to Redis using URL: ${config.redis.url}`);
      }

      this.client = createClient(redisConfig);

      this.client.on('error', (error) => {
        logger.error('Redis client error:', error);
        this.isConnected = false;
      });

      this.client.on('connect', () => {
        logger.info('Redis client connected');
        this.isConnected = true;
      });

      this.client.on('disconnect', () => {
        logger.warn('Redis client disconnected');
        this.isConnected = false;
      });

      await this.client.connect();
      logger.info('✅ Redis connected successfully');
    } catch (error) {
      logger.error('Failed to connect to Redis:', error);
      this.isConnected = false;
      // Don't throw error - allow app to continue without Redis
    }
  }

  public async disconnect(): Promise<void> {
    if (this.client && this.isConnected) {
      try {
        await this.client.quit();
        this.isConnected = false;
        logger.info('Redis disconnected');
      } catch (error) {
        logger.error('Error disconnecting from Redis:', error);
      }
    }
  }

  public getClient(): RedisClientType | null {
    return this.client;
  }

  public isReady(): boolean {
    return this.isConnected && this.client !== null;
  }

  // Cache operations
  public async set<T>(key: string, value: T, ttlSeconds?: number): Promise<boolean> {
    if (!this.isReady()) {
      logger.warn('Redis not available, skipping cache set');
      return false;
    }

    try {
      const cacheEntry: CacheEntry<T> = {
        data: value,
        timestamp: Date.now(),
        ttl: ttlSeconds || 300, // Default 5 minutes
      };

      const serialized = JSON.stringify(cacheEntry);

      if (ttlSeconds) {
        await this.client!.setEx(key, ttlSeconds, serialized);
      } else {
        await this.client!.set(key, serialized);
      }

      return true;
    } catch (error) {
      logger.error(`Error setting cache key ${key}:`, error);
      return false;
    }
  }

  public async get<T>(key: string): Promise<T | null> {
    if (!this.isReady()) {
      return null;
    }

    try {
      const cached = await this.client!.get(key);

      if (!cached) {
        return null;
      }

      const cacheEntry: CacheEntry<T> = JSON.parse(cached);

      // Check if cache entry has expired
      const now = Date.now();
      const age = (now - cacheEntry.timestamp) / 1000; // age in seconds

      if (age > cacheEntry.ttl) {
        // Cache expired, delete it
        await this.delete(key);
        return null;
      }

      return cacheEntry.data;
    } catch (error) {
      logger.error(`Error getting cache key ${key}:`, error);
      return null;
    }
  }

  public async delete(key: string): Promise<boolean> {
    if (!this.isReady()) {
      return false;
    }

    try {
      await this.client!.del(key);
      return true;
    } catch (error) {
      logger.error(`Error deleting cache key ${key}:`, error);
      return false;
    }
  }

  public async exists(key: string): Promise<boolean> {
    if (!this.isReady()) {
      return false;
    }

    try {
      const result = await this.client!.exists(key);
      return result === 1;
    } catch (error) {
      logger.error(`Error checking cache key existence ${key}:`, error);
      return false;
    }
  }

  public async deletePattern(pattern: string): Promise<number> {
    if (!this.isReady()) {
      return 0;
    }

    try {
      const keys = await this.client!.keys(pattern);
      if (keys.length === 0) {
        return 0;
      }

      await this.client!.del(keys);
      return keys.length;
    } catch (error) {
      logger.error(`Error deleting cache pattern ${pattern}:`, error);
      return 0;
    }
  }

  public async flushAll(): Promise<boolean> {
    if (!this.isReady()) {
      return false;
    }

    try {
      await this.client!.flushAll();
      return true;
    } catch (error) {
      logger.error('Error flushing all cache:', error);
      return false;
    }
  }

  // Utility methods for common cache patterns
  public async getOrSet<T>(
    key: string,
    fetchFunction: () => Promise<T>,
    ttlSeconds?: number
  ): Promise<T> {
    // Try to get from cache first
    const cached = await this.get<T>(key);
    if (cached !== null) {
      return cached;
    }

    // If not in cache, fetch the data
    const data = await fetchFunction();

    // Store in cache for next time
    await this.set(key, data, ttlSeconds);

    return data;
  }

  // Generate cache keys
  public static generateKey(prefix: string, ...parts: (string | number)[]): string {
    return `${prefix}:${parts.join(':')}`;
  }

  // Additional Redis operations for job queue
  public async zadd(key: string, score: number, member: string): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.zAdd(key, { score, value: member });
  }

  public async zrangebyscore(key: string, min: number, max: number): Promise<string[]> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.zRangeByScore(key, min, max);
  }

  public async zrem(key: string, member: string): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.zRem(key, member);
  }

  public async zcard(key: string): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.zCard(key);
  }

  public async lpush(key: string, ...values: string[]): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.lPush(key, values);
  }

  public async rpop(key: string): Promise<string | null> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.rPop(key);
  }

  public async llen(key: string): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.lLen(key);
  }

  public async sadd(key: string, ...members: string[]): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.sAdd(key, members);
  }

  public async srem(key: string, ...members: string[]): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.sRem(key, members);
  }

  public async scard(key: string): Promise<number> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.sCard(key);
  }

  public async keys(pattern: string): Promise<string[]> {
    if (!this.isReady()) {
      throw new Error('Redis not connected');
    }
    return await this.client!.keys(pattern);
  }

  // Health check
  public async healthCheck(): Promise<boolean> {
    if (!this.isReady()) {
      return false;
    }

    try {
      await this.client!.ping();
      return true;
    } catch (error) {
      logger.error('Redis health check failed:', error);
      return false;
    }
  }
}
