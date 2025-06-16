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
      this.client = createClient({
        url: config.redis.url,
        password: config.redis.password,
        socket: {
          reconnectStrategy: (retries) => {
            if (retries > 10) {
              logger.error('Redis reconnection failed after 10 attempts');
              return false;
            }
            return Math.min(retries * 100, 3000);
          },
        },
      });

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
}
