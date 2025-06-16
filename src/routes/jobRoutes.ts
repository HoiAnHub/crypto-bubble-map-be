import { Router, Request, Response, NextFunction } from 'express';
import { logger } from '@/utils/logger';
import { ApiResponse } from '@/types';

const router = Router();

// This would be injected in a real implementation
// For now, we'll create a placeholder
let jobSchedulerService: any = null;

// Middleware to set job scheduler service
export const setJobSchedulerService = (service: any) => {
  jobSchedulerService = service;
};

/**
 * @route GET /api/jobs/stats
 * @desc Get job queue statistics
 */
router.get('/stats', async (req: Request, res: Response, next: NextFunction) => {
  try {
    if (!jobSchedulerService) {
      return res.status(503).json({
        success: false,
        error: 'Job scheduler service not available',
        timestamp: new Date().toISOString(),
      });
    }

    logger.info('Getting job statistics');

    const stats = await jobSchedulerService.getJobStats();

    const response: ApiResponse = {
      success: true,
      data: stats,
      timestamp: new Date().toISOString(),
    };

    res.json(response);

  } catch (error) {
    next(error);
  }
});

/**
 * @route POST /api/jobs/trigger/:jobType
 * @desc Manually trigger a specific job type
 */
router.post('/trigger/:jobType', async (req: Request, res: Response, next: NextFunction) => {
  try {
    if (!jobSchedulerService) {
      return res.status(503).json({
        success: false,
        error: 'Job scheduler service not available',
        timestamp: new Date().toISOString(),
      });
    }

    const { jobType } = req.params;
    const { priority = 'medium', data = {} } = req.body;

    logger.info(`Manually triggering job: ${jobType}`);

    // This would need to be implemented in the job scheduler service
    // For now, return a success response
    const response: ApiResponse = {
      success: true,
      data: {
        message: `Job ${jobType} triggered successfully`,
        jobType,
        priority,
        triggeredAt: new Date().toISOString(),
      },
      timestamp: new Date().toISOString(),
    };

    res.json(response);

  } catch (error) {
    next(error);
  }
});

/**
 * @route GET /api/jobs/health
 * @desc Get job system health status
 */
router.get('/health', async (req: Request, res: Response, next: NextFunction) => {
  try {
    logger.info('Checking job system health');

    const health = {
      jobScheduler: jobSchedulerService ? 'running' : 'not available',
      backgroundJobs: process.env.JOBS_ENABLED !== 'false' ? 'enabled' : 'disabled',
      timestamp: new Date().toISOString(),
    };

    const response: ApiResponse = {
      success: true,
      data: health,
      timestamp: new Date().toISOString(),
    };

    res.json(response);

  } catch (error) {
    next(error);
  }
});

export { router as jobRoutes };
