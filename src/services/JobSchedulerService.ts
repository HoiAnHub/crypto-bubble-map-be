import * as cron from 'node-cron';
import { logger } from '@/utils/logger';
import { JobQueueService, JobType, JobPriority } from './JobQueueService';
import { RedisService } from './RedisService';
import { DatabaseService } from './DatabaseService';
import { EthereumService } from './EthereumService';
import { Neo4jService } from './Neo4jService';
import { MarketDataCrawler } from './crawlers/MarketDataCrawler';
import { WalletDataCrawler } from './crawlers/WalletDataCrawler';
import { config } from '@/config/config';

export class JobSchedulerService {
  private jobQueueService: JobQueueService;
  private redisService: RedisService;
  private databaseService: DatabaseService;
  private ethereumService: EthereumService;
  private neo4jService: Neo4jService;
  private marketDataCrawler: MarketDataCrawler;
  private walletDataCrawler: WalletDataCrawler;

  private scheduledJobs: Map<string, cron.ScheduledTask> = new Map();
  private isProcessing = false;
  private processingInterval?: NodeJS.Timeout;

  constructor(
    jobQueueService: JobQueueService,
    redisService: RedisService,
    databaseService: DatabaseService,
    ethereumService: EthereumService,
    neo4jService: Neo4jService
  ) {
    this.jobQueueService = jobQueueService;
    this.redisService = redisService;
    this.databaseService = databaseService;
    this.ethereumService = ethereumService;
    this.neo4jService = neo4jService;

    // Initialize crawlers
    this.marketDataCrawler = new MarketDataCrawler(redisService, databaseService);
    this.walletDataCrawler = new WalletDataCrawler(
      redisService,
      databaseService,
      ethereumService,
      neo4jService
    );
  }

  /**
   * Start the job scheduler
   */
  public async start(): Promise<void> {
    if (!config.jobs.enabled) {
      logger.info('Background jobs are disabled');
      return;
    }

    logger.info('Starting job scheduler service');

    try {
      // Schedule all background jobs
      this.scheduleMarketDataJob();
      this.schedulePopularWalletsJob();
      this.scheduleNetworkStatsJob();
      this.scheduleWalletRefreshJob();
      this.scheduleCleanupJob();

      // Start job processor
      this.startJobProcessor();

      logger.info('Job scheduler started successfully');

    } catch (error) {
      logger.error('Error starting job scheduler:', error);
      throw error;
    }
  }

  /**
   * Stop the job scheduler
   */
  public async stop(): Promise<void> {
    logger.info('Stopping job scheduler service');

    // Stop all scheduled jobs
    for (const [name, task] of this.scheduledJobs) {
      task.stop();
      logger.info(`Stopped scheduled job: ${name}`);
    }
    this.scheduledJobs.clear();

    // Stop job processor
    if (this.processingInterval) {
      clearInterval(this.processingInterval);
      this.processingInterval = undefined;
    }
    this.isProcessing = false;

    logger.info('Job scheduler stopped');
  }

  /**
   * Schedule market data crawling job
   */
  private scheduleMarketDataJob(): void {
    const task = cron.schedule(config.jobs.intervals.marketData, async () => {
      logger.info('Scheduling market data crawl job');

      await this.jobQueueService.addJob(
        JobType.MARKET_DATA_CRAWL,
        {},
        JobPriority.HIGH
      );
    }, {
      scheduled: false,
      timezone: 'UTC'
    });

    task.start();
    this.scheduledJobs.set('market_data', task);
    logger.info(`Market data job scheduled: ${config.jobs.intervals.marketData}`);
  }

  /**
   * Schedule popular wallets discovery job
   */
  private schedulePopularWalletsJob(): void {
    const task = cron.schedule(config.jobs.intervals.popularWallets, async () => {
      logger.info('Scheduling popular wallets discovery job');

      await this.jobQueueService.addJob(
        JobType.POPULAR_WALLETS_DISCOVERY,
        {},
        JobPriority.MEDIUM
      );
    }, {
      scheduled: false,
      timezone: 'UTC'
    });

    task.start();
    this.scheduledJobs.set('popular_wallets', task);
    logger.info(`Popular wallets job scheduled: ${config.jobs.intervals.popularWallets}`);
  }

  /**
   * Schedule network statistics job
   */
  private scheduleNetworkStatsJob(): void {
    const task = cron.schedule(config.jobs.intervals.networkStats, async () => {
      logger.info('Scheduling network stats job');

      await this.jobQueueService.addJob(
        JobType.NETWORK_STATS_CRAWL,
        {},
        JobPriority.MEDIUM
      );
    }, {
      scheduled: false,
      timezone: 'UTC'
    });

    task.start();
    this.scheduledJobs.set('network_stats', task);
    logger.info(`Network stats job scheduled: ${config.jobs.intervals.networkStats}`);
  }

  /**
   * Schedule wallet refresh job
   */
  private scheduleWalletRefreshJob(): void {
    const task = cron.schedule(config.jobs.intervals.walletRefresh, async () => {
      logger.info('Scheduling wallet refresh job');

      // Get list of wallets that need refreshing
      const walletsToRefresh = await this.getWalletsForRefresh();

      if (walletsToRefresh.length > 0) {
        // Split into batches
        const batchSize = config.jobs.batchSizes.walletBatch;
        for (let i = 0; i < walletsToRefresh.length; i += batchSize) {
          const batch = walletsToRefresh.slice(i, i + batchSize);

          await this.jobQueueService.addJob(
            JobType.WALLET_REFRESH,
            { addresses: batch },
            JobPriority.LOW
          );
        }
      }
    }, {
      scheduled: false,
      timezone: 'UTC'
    });

    task.start();
    this.scheduledJobs.set('wallet_refresh', task);
    logger.info(`Wallet refresh job scheduled: ${config.jobs.intervals.walletRefresh}`);
  }

  /**
   * Schedule cleanup job
   */
  private scheduleCleanupJob(): void {
    const task = cron.schedule(config.jobs.intervals.cleanup, async () => {
      logger.info('Scheduling cleanup job');

      await this.jobQueueService.addJob(
        JobType.DATA_CLEANUP,
        {},
        JobPriority.LOW
      );
    }, {
      scheduled: false,
      timezone: 'UTC'
    });

    task.start();
    this.scheduledJobs.set('cleanup', task);
    logger.info(`Cleanup job scheduled: ${config.jobs.intervals.cleanup}`);
  }

  /**
   * Start the job processor
   */
  private startJobProcessor(): void {
    if (this.isProcessing) {
      return;
    }

    this.isProcessing = true;

    this.processingInterval = setInterval(async () => {
      try {
        await this.processNextJob();
      } catch (error) {
        logger.error('Error in job processor:', error);
      }
    }, 5000); // Check for jobs every 5 seconds

    logger.info('Job processor started');
  }

  /**
   * Process the next job in the queue
   */
  private async processNextJob(): Promise<void> {
    const job = await this.jobQueueService.getNextJob();

    if (!job) {
      return; // No jobs to process
    }

    logger.info(`Processing job ${job.id} (${job.type})`);

    try {
      await this.jobQueueService.startJob(job.id);

      let result: any;

      switch (job.type) {
        case JobType.MARKET_DATA_CRAWL:
          result = await this.marketDataCrawler.crawlMarketData();
          break;

        case JobType.POPULAR_WALLETS_DISCOVERY:
          result = await this.walletDataCrawler.discoverPopularWallets();
          break;

        case JobType.WALLET_DATA_CRAWL:
          result = await this.walletDataCrawler.crawlWalletBatch(job.data.addresses || []);
          break;

        case JobType.WALLET_REFRESH:
          result = await this.walletDataCrawler.refreshWalletData(job.data.addresses || []);
          break;

        case JobType.NETWORK_STATS_CRAWL:
          result = await this.processNetworkStatsJob();
          break;

        case JobType.DATA_CLEANUP:
          result = await this.processCleanupJob();
          break;

        default:
          throw new Error(`Unknown job type: ${job.type}`);
      }

      await this.jobQueueService.completeJob(job.id, result);
      logger.info(`Job ${job.id} completed successfully`);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      await this.jobQueueService.failJob(job.id, errorMessage);
      logger.error(`Job ${job.id} failed: ${errorMessage}`);
    }
  }

  /**
   * Process network statistics job
   */
  private async processNetworkStatsJob(): Promise<any> {
    logger.info('Processing network stats job');

    try {
      const blockNumber = await this.ethereumService.getCurrentBlockNumber();
      const ethPrice = await this.ethereumService.getETHPrice();

      const networkStats = {
        blockNumber,
        ethPrice,
        timestamp: new Date(),
      };

      // Cache network stats
      await this.redisService.set('network_stats', networkStats, 600); // 10 minutes TTL

      return networkStats;

    } catch (error) {
      logger.error('Error processing network stats job:', error);
      throw error;
    }
  }

  /**
   * Process cleanup job
   */
  private async processCleanupJob(): Promise<any> {
    logger.info('Processing cleanup job');

    try {
      // Clean up old jobs
      const cleanedJobs = await this.jobQueueService.cleanupOldJobs(24);

      // Clean up old cache entries
      const cleanedCache = await this.cleanupOldCacheEntries();

      // Clean up old market data history (keep last 30 days)
      const cleanedMarketData = await this.cleanupOldMarketData();

      const result = {
        cleanedJobs,
        cleanedCache,
        cleanedMarketData,
        timestamp: new Date(),
      };

      logger.info(`Cleanup completed: ${JSON.stringify(result)}`);
      return result;

    } catch (error) {
      logger.error('Error processing cleanup job:', error);
      throw error;
    }
  }

  /**
   * Get wallets that need refreshing
   */
  private async getWalletsForRefresh(): Promise<string[]> {
    try {
      // Get high priority wallets
      const highPriorityWallets = config.jobs.priorities.high;

      // Get recently active wallets from database
      const result = await this.databaseService.query(
        `
        SELECT address FROM wallets
        WHERE last_activity > NOW() - INTERVAL '7 days'
          AND updated_at < NOW() - INTERVAL '2 hours'
        ORDER BY transaction_count DESC, balance_usd DESC
        LIMIT 100
        `,
        []
      );

      const activeWallets = result.rows.map(row => row.address);

      // Combine and deduplicate
      const allWallets = [...new Set([...highPriorityWallets, ...activeWallets])];

      logger.info(`Found ${allWallets.length} wallets for refresh`);
      return allWallets;

    } catch (error) {
      logger.error('Error getting wallets for refresh:', error);
      return config.jobs.priorities.high; // Fallback to high priority wallets
    }
  }

  /**
   * Clean up old cache entries
   */
  private async cleanupOldCacheEntries(): Promise<number> {
    // This would implement cache cleanup logic
    // For now, return 0
    return 0;
  }

  /**
   * Clean up old market data history
   */
  private async cleanupOldMarketData(): Promise<number> {
    try {
      const result = await this.databaseService.query(
        'DELETE FROM market_data_history WHERE created_at < NOW() - INTERVAL \'30 days\'',
        []
      );

      return result.rowCount || 0;

    } catch (error) {
      logger.error('Error cleaning up old market data:', error);
      return 0;
    }
  }

  /**
   * Get job statistics
   */
  public async getJobStats(): Promise<any> {
    const queueStats = await this.jobQueueService.getJobStats();

    return {
      ...queueStats,
      scheduledJobs: Array.from(this.scheduledJobs.keys()),
      isProcessing: this.isProcessing,
      timestamp: new Date(),
    };
  }
}
