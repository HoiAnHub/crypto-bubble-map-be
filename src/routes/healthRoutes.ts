import { Router, Request, Response } from 'express';
import { DatabaseService } from '@/services/DatabaseService';
import { RedisService } from '@/services/RedisService';
import { Neo4jService } from '@/services/Neo4jService';
import { EthereumService } from '@/services/EthereumService';
import { logger } from '@/utils/logger';

const router = Router();

/**
 * @route GET /health
 * @desc Basic health check
 */
router.get('/', (req: Request, res: Response) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    version: '1.0.0',
  });
});

/**
 * @route GET /health/detailed
 * @desc Detailed health check including all services
 */
router.get('/detailed', async (req: Request, res: Response) => {
  const healthChecks = {
    api: true,
    database: false,
    redis: false,
    neo4j: false,
    ethereum: false,
  };

  const details: any = {
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    version: '1.0.0',
    services: {},
  };

  try {
    // Check PostgreSQL
    const dbService = new DatabaseService();
    try {
      await dbService.connect();
      healthChecks.database = await dbService.healthCheck();
      details.services.database = {
        status: healthChecks.database ? 'healthy' : 'unhealthy',
        message: healthChecks.database ? 'Connected' : 'Connection failed',
      };
      await dbService.disconnect();
    } catch (error) {
      details.services.database = {
        status: 'unhealthy',
        message: error instanceof Error ? error.message : 'Unknown error',
      };
    }

    // Check Redis
    const redisService = new RedisService();
    try {
      await redisService.connect();
      healthChecks.redis = redisService.isReady();
      details.services.redis = {
        status: healthChecks.redis ? 'healthy' : 'unhealthy',
        message: healthChecks.redis ? 'Connected' : 'Connection failed',
      };
      await redisService.disconnect();
    } catch (error) {
      details.services.redis = {
        status: 'unhealthy',
        message: error instanceof Error ? error.message : 'Unknown error',
      };
    }

    // Check Neo4j
    const neo4jService = new Neo4jService();
    try {
      await neo4jService.connect();
      healthChecks.neo4j = await neo4jService.healthCheck();
      details.services.neo4j = {
        status: healthChecks.neo4j ? 'healthy' : 'unhealthy',
        message: healthChecks.neo4j ? 'Connected' : 'Connection failed',
      };
      await neo4jService.disconnect();
    } catch (error) {
      details.services.neo4j = {
        status: 'unhealthy',
        message: error instanceof Error ? error.message : 'Unknown error',
      };
    }

    // Check Ethereum connection
    const ethereumService = new EthereumService();
    try {
      const blockNumber = await ethereumService.getCurrentBlockNumber();
      healthChecks.ethereum = blockNumber > 0;
      details.services.ethereum = {
        status: healthChecks.ethereum ? 'healthy' : 'unhealthy',
        message: healthChecks.ethereum ? `Connected (block: ${blockNumber})` : 'Connection failed',
        currentBlock: blockNumber,
      };
    } catch (error) {
      details.services.ethereum = {
        status: 'unhealthy',
        message: error instanceof Error ? error.message : 'Unknown error',
      };
    }

    // Overall health status
    const overallHealth = Object.values(healthChecks).every(status => status);
    
    const response = {
      status: overallHealth ? 'healthy' : 'degraded',
      overall: overallHealth,
      ...details,
    };

    // Return appropriate HTTP status
    const httpStatus = overallHealth ? 200 : 503;
    res.status(httpStatus).json(response);

  } catch (error) {
    logger.error('Health check error:', error);
    res.status(500).json({
      status: 'unhealthy',
      error: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
    });
  }
});

/**
 * @route GET /health/ready
 * @desc Readiness probe for Kubernetes
 */
router.get('/ready', async (req: Request, res: Response) => {
  try {
    // Check if essential services are ready
    const dbService = new DatabaseService();
    await dbService.connect();
    const dbReady = await dbService.healthCheck();
    await dbService.disconnect();

    if (dbReady) {
      res.json({
        status: 'ready',
        timestamp: new Date().toISOString(),
      });
    } else {
      res.status(503).json({
        status: 'not ready',
        message: 'Database not ready',
        timestamp: new Date().toISOString(),
      });
    }
  } catch (error) {
    res.status(503).json({
      status: 'not ready',
      error: error instanceof Error ? error.message : 'Unknown error',
      timestamp: new Date().toISOString(),
    });
  }
});

/**
 * @route GET /health/live
 * @desc Liveness probe for Kubernetes
 */
router.get('/live', (req: Request, res: Response) => {
  res.json({
    status: 'alive',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  });
});

export { router as healthRoutes };
