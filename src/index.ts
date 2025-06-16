import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import compression from 'compression';
import dotenv from 'dotenv';

import { config } from '@/config/config';
import { logger } from '@/utils/logger';
import { errorHandler } from '@/middleware/errorHandler';
import { rateLimiter } from '@/middleware/rateLimiter';
import { walletRoutes } from '@/routes/walletRoutes';
import { healthRoutes } from '@/routes/healthRoutes';
import { DatabaseService } from '@/services/DatabaseService';
import { RedisService } from '@/services/RedisService';

// Load environment variables
dotenv.config();

class App {
  public app: express.Application;
  private databaseService: DatabaseService;
  private redisService: RedisService;

  constructor() {
    this.app = express();
    this.databaseService = new DatabaseService();
    this.redisService = new RedisService();
    
    this.initializeMiddleware();
    this.initializeRoutes();
    this.initializeErrorHandling();
  }

  private initializeMiddleware(): void {
    // Security middleware
    this.app.use(helmet());
    
    // CORS configuration
    this.app.use(cors({
      origin: config.corsOrigin,
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization', 'X-API-Key']
    }));

    // Compression middleware
    this.app.use(compression());

    // Request logging
    this.app.use(morgan('combined', {
      stream: { write: (message) => logger.info(message.trim()) }
    }));

    // Body parsing middleware
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true, limit: '10mb' }));

    // Rate limiting
    this.app.use(rateLimiter);
  }

  private initializeRoutes(): void {
    // Health check routes
    this.app.use('/health', healthRoutes);
    
    // API routes
    this.app.use('/api/wallets', walletRoutes);

    // Root endpoint
    this.app.get('/', (req, res) => {
      res.json({
        message: 'Crypto Bubble Map Backend API',
        version: '1.0.0',
        status: 'running',
        timestamp: new Date().toISOString()
      });
    });

    // 404 handler
    this.app.use('*', (req, res) => {
      res.status(404).json({
        error: 'Not Found',
        message: `Route ${req.originalUrl} not found`,
        timestamp: new Date().toISOString()
      });
    });
  }

  private initializeErrorHandling(): void {
    this.app.use(errorHandler);
  }

  public async start(): Promise<void> {
    try {
      // Initialize database connections
      await this.databaseService.connect();
      await this.redisService.connect();

      // Start the server
      const port = config.port;
      const host = config.host;

      this.app.listen(port, host, () => {
        logger.info(`🚀 Server running on http://${host}:${port}`);
        logger.info(`📊 Environment: ${config.nodeEnv}`);
        logger.info(`🔗 Ethereum Network: ${config.ethereum.network}`);
      });

      // Graceful shutdown handling
      this.setupGracefulShutdown();

    } catch (error) {
      logger.error('Failed to start server:', error);
      process.exit(1);
    }
  }

  private setupGracefulShutdown(): void {
    const gracefulShutdown = async (signal: string) => {
      logger.info(`Received ${signal}. Starting graceful shutdown...`);
      
      try {
        await this.databaseService.disconnect();
        await this.redisService.disconnect();
        logger.info('Graceful shutdown completed');
        process.exit(0);
      } catch (error) {
        logger.error('Error during graceful shutdown:', error);
        process.exit(1);
      }
    };

    process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
    process.on('SIGINT', () => gracefulShutdown('SIGINT'));
  }
}

// Start the application
const app = new App();
app.start().catch((error) => {
  logger.error('Failed to start application:', error);
  process.exit(1);
});

export default app;
