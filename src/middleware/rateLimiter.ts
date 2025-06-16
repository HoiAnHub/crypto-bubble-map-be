import { Request, Response, NextFunction } from 'express';
import { RateLimiterRedis } from 'rate-limiter-flexible';
import { config } from '@/config/config';
import { RedisService } from '@/services/RedisService';
import { RateLimitError } from './errorHandler';

class RateLimiterService {
  private rateLimiter: RateLimiterRedis | null = null;
  private redisService: RedisService;

  constructor() {
    this.redisService = new RedisService();
    this.initializeRateLimiter();
  }

  private async initializeRateLimiter(): Promise<void> {
    try {
      await this.redisService.connect();
      const redisClient = this.redisService.getClient();

      if (redisClient) {
        this.rateLimiter = new RateLimiterRedis({
          storeClient: redisClient,
          keyPrefix: 'rl_api',
          points: config.rateLimiting.maxRequests, // Number of requests
          duration: Math.floor(config.rateLimiting.windowMs / 1000), // Per duration in seconds
          blockDuration: 10, // Block for 10 seconds if limit exceeded (reduced from 60)
          execEvenly: false, // Don't spread requests evenly (allow bursts)
        });
      }
    } catch (error) {
      console.error('Failed to initialize rate limiter:', error);
      // Fallback to in-memory rate limiter if Redis is not available
      this.initializeFallbackRateLimiter();
    }
  }

  private initializeFallbackRateLimiter(): void {
    const { RateLimiterMemory } = require('rate-limiter-flexible');

    this.rateLimiter = new RateLimiterMemory({
      keyPrefix: 'rl_api_memory',
      points: config.rateLimiting.maxRequests,
      duration: Math.floor(config.rateLimiting.windowMs / 1000),
      blockDuration: 60,
      execEvenly: true,
    });
  }

  public async checkRateLimit(req: Request): Promise<void> {
    if (!this.rateLimiter) {
      // If rate limiter is not initialized, allow the request
      return;
    }

    const key = this.getClientKey(req);

    try {
      await this.rateLimiter.consume(key);
    } catch (rejRes: any) {
      // Rate limit exceeded
      const remainingPoints = rejRes.remainingPoints || 0;
      const msBeforeNext = rejRes.msBeforeNext || 0;
      const totalHits = rejRes.totalHits || 0;

      throw new RateLimitError(
        `Rate limit exceeded. Try again in ${Math.round(msBeforeNext / 1000)} seconds.`
      );
    }
  }

  private getClientKey(req: Request): string {
    // Use IP address as the primary identifier
    const ip = req.ip || req.connection.remoteAddress || 'unknown';

    // If API key is provided, use it for more specific rate limiting
    const apiKey = req.headers['x-api-key'] as string;
    if (apiKey) {
      return `api_key:${apiKey}`;
    }

    // Use IP address
    return `ip:${ip}`;
  }

  public async getRateLimitInfo(req: Request): Promise<{
    limit: number;
    remaining: number;
    reset: Date;
  }> {
    if (!this.rateLimiter) {
      return {
        limit: config.rateLimiting.maxRequests,
        remaining: config.rateLimiting.maxRequests,
        reset: new Date(Date.now() + config.rateLimiting.windowMs),
      };
    }

    const key = this.getClientKey(req);

    try {
      const resRateLimiter = await this.rateLimiter.get(key);

      if (resRateLimiter) {
        return {
          limit: config.rateLimiting.maxRequests,
          remaining: Math.max(0, config.rateLimiting.maxRequests - (resRateLimiter as any).totalHits),
          reset: new Date(Date.now() + resRateLimiter.msBeforeNext),
        };
      } else {
        return {
          limit: config.rateLimiting.maxRequests,
          remaining: config.rateLimiting.maxRequests,
          reset: new Date(Date.now() + config.rateLimiting.windowMs),
        };
      }
    } catch (error) {
      // Return default values if there's an error
      return {
        limit: config.rateLimiting.maxRequests,
        remaining: config.rateLimiting.maxRequests,
        reset: new Date(Date.now() + config.rateLimiting.windowMs),
      };
    }
  }
}

// Create singleton instance
const rateLimiterService = new RateLimiterService();

// Express middleware
export const rateLimiter = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
  try {
    // Skip rate limiting in development mode
    if (config.nodeEnv === 'development') {
      next();
      return;
    }

    await rateLimiterService.checkRateLimit(req);

    // Add rate limit headers
    const rateLimitInfo = await rateLimiterService.getRateLimitInfo(req);
    res.set({
      'X-RateLimit-Limit': rateLimitInfo.limit.toString(),
      'X-RateLimit-Remaining': rateLimitInfo.remaining.toString(),
      'X-RateLimit-Reset': rateLimitInfo.reset.toISOString(),
    });

    next();
  } catch (error) {
    next(error);
  }
};

export { rateLimiterService };
