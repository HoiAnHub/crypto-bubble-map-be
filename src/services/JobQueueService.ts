import { RedisService } from './RedisService';
import { logger } from '@/utils/logger';
import { config } from '@/config/config';

export interface Job {
  id: string;
  type: JobType;
  data: any;
  priority: JobPriority;
  attempts: number;
  maxAttempts: number;
  createdAt: Date;
  scheduledAt?: Date;
  startedAt?: Date;
  completedAt?: Date;
  failedAt?: Date;
  error?: string;
  result?: any;
}

export enum JobType {
  MARKET_DATA_CRAWL = 'market_data_crawl',
  WALLET_DATA_CRAWL = 'wallet_data_crawl',
  NETWORK_STATS_CRAWL = 'network_stats_crawl',
  POPULAR_WALLETS_DISCOVERY = 'popular_wallets_discovery',
  WALLET_NETWORK_BUILD = 'wallet_network_build',
  DATA_CLEANUP = 'data_cleanup',
  WALLET_REFRESH = 'wallet_refresh',
}

export enum JobPriority {
  LOW = 0,
  MEDIUM = 1,
  HIGH = 2,
  CRITICAL = 3,
}

export enum JobStatus {
  PENDING = 'pending',
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  FAILED = 'failed',
  RETRYING = 'retrying',
}

export class JobQueueService {
  private redisService: RedisService;
  private isProcessing = false;
  private processingInterval?: NodeJS.Timeout;

  constructor(redisService: RedisService) {
    this.redisService = redisService;
  }

  /**
   * Add a job to the queue
   */
  public async addJob(
    type: JobType,
    data: any,
    priority: JobPriority = JobPriority.MEDIUM,
    scheduledAt?: Date
  ): Promise<string> {
    const jobId = this.generateJobId();

    const job: Job = {
      id: jobId,
      type,
      data,
      priority,
      attempts: 0,
      maxAttempts: config.jobs.retries.maxRetries,
      createdAt: new Date(),
      scheduledAt,
    };

    // Store job data
    await this.redisService.set(`job:${jobId}`, job, 86400); // 24 hours TTL

    // Add to appropriate queue based on priority and scheduling
    if (scheduledAt && scheduledAt > new Date()) {
      // Scheduled job
      await this.redisService.zadd('jobs:scheduled', scheduledAt.getTime(), jobId);
    } else {
      // Immediate job - add to priority queue
      const queueKey = this.getQueueKey(priority);
      await this.redisService.lpush(queueKey, jobId);
    }

    logger.info(`Job ${jobId} (${type}) added to queue with priority ${priority}`);
    return jobId;
  }

  /**
   * Get the next job to process
   */
  public async getNextJob(): Promise<Job | null> {
    // First, check for scheduled jobs that are ready
    await this.moveScheduledJobsToQueue();

    // Get job from priority queues (highest priority first)
    const priorities = [JobPriority.CRITICAL, JobPriority.HIGH, JobPriority.MEDIUM, JobPriority.LOW];

    for (const priority of priorities) {
      const queueKey = this.getQueueKey(priority);
      const jobId = await this.redisService.rpop(queueKey);

      if (jobId) {
        const job = await this.redisService.get<Job>(`job:${jobId}`);
        if (job) {
          return job;
        }
      }
    }

    return null;
  }

  /**
   * Mark job as started
   */
  public async startJob(jobId: string): Promise<void> {
    const job = await this.redisService.get<Job>(`job:${jobId}`);
    if (job) {
      job.startedAt = new Date();
      job.attempts += 1;
      await this.redisService.set(`job:${jobId}`, job, 86400);

      // Add to processing set
      await this.redisService.sadd('jobs:processing', jobId);
    }
  }

  /**
   * Mark job as completed
   */
  public async completeJob(jobId: string, result?: any): Promise<void> {
    const job = await this.redisService.get<Job>(`job:${jobId}`);
    if (job) {
      job.completedAt = new Date();
      job.result = result;
      await this.redisService.set(`job:${jobId}`, job, 86400);

      // Remove from processing set
      await this.redisService.srem('jobs:processing', jobId);

      logger.info(`Job ${jobId} (${job.type}) completed successfully`);
    }
  }

  /**
   * Mark job as failed
   */
  public async failJob(jobId: string, error: string): Promise<void> {
    const job = await this.redisService.get<Job>(`job:${jobId}`);
    if (!job) return;

    job.error = error;
    job.failedAt = new Date();

    // Check if we should retry
    if (job.attempts < job.maxAttempts) {
      // Schedule retry with exponential backoff
      const retryDelay = config.jobs.retries.retryDelay * Math.pow(2, job.attempts - 1);
      const retryAt = new Date(Date.now() + retryDelay);

      job.scheduledAt = retryAt;
      await this.redisService.set(`job:${jobId}`, job, 86400);
      await this.redisService.zadd('jobs:scheduled', retryAt.getTime(), jobId);

      logger.warn(`Job ${jobId} (${job.type}) failed, retrying in ${retryDelay}ms (attempt ${job.attempts}/${job.maxAttempts})`);
    } else {
      // Max retries reached
      await this.redisService.set(`job:${jobId}`, job, 86400);
      logger.error(`Job ${jobId} (${job.type}) failed permanently after ${job.attempts} attempts: ${error}`);
    }

    // Remove from processing set
    await this.redisService.srem('jobs:processing', jobId);
  }

  /**
   * Get job statistics
   */
  public async getJobStats(): Promise<{
    pending: number;
    processing: number;
    scheduled: number;
    queueSizes: Record<string, number>;
  }> {
    const [processing, scheduled] = await Promise.all([
      this.redisService.scard('jobs:processing'),
      this.redisService.zcard('jobs:scheduled'),
    ]);

    const queueSizes: Record<string, number> = {};
    for (const priority of Object.values(JobPriority)) {
      if (typeof priority === 'number') {
        const queueKey = this.getQueueKey(priority);
        queueSizes[JobPriority[priority]] = await this.redisService.llen(queueKey);
      }
    }

    const pending = Object.values(queueSizes).reduce((sum, size) => sum + size, 0);

    return {
      pending,
      processing,
      scheduled,
      queueSizes,
    };
  }

  /**
   * Clean up old completed jobs
   */
  public async cleanupOldJobs(olderThanHours: number = 24): Promise<number> {
    const cutoffTime = Date.now() - (olderThanHours * 60 * 60 * 1000);
    const pattern = 'job:*';

    let cleaned = 0;
    const keys = await this.redisService.keys(pattern);

    for (const key of keys) {
      const job = await this.redisService.get<Job>(key);
      if (job && job.completedAt && new Date(job.completedAt).getTime() < cutoffTime) {
        await this.redisService.delete(key);
        cleaned++;
      }
    }

    logger.info(`Cleaned up ${cleaned} old jobs`);
    return cleaned;
  }

  /**
   * Move scheduled jobs that are ready to the appropriate queues
   */
  private async moveScheduledJobsToQueue(): Promise<void> {
    const now = Date.now();
    const readyJobs = await this.redisService.zrangebyscore('jobs:scheduled', 0, now);

    for (const jobId of readyJobs) {
      const job = await this.redisService.get<Job>(`job:${jobId}`);
      if (job) {
        // Remove from scheduled set
        await this.redisService.zrem('jobs:scheduled', jobId);

        // Add to priority queue
        const queueKey = this.getQueueKey(job.priority);
        await this.redisService.lpush(queueKey, jobId);
      }
    }
  }

  private getQueueKey(priority: JobPriority): string {
    return `jobs:queue:${JobPriority[priority].toLowerCase()}`;
  }

  private generateJobId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }
}
