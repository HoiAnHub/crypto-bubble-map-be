import { Request, Response, NextFunction } from 'express';
import Joi from 'joi';
import { WalletService } from '@/services/WalletService';
import { logger } from '@/utils/logger';
import { ValidationError, NotFoundError } from '@/middleware/errorHandler';
import {
  ApiResponse,
  WalletNetworkRequest,
  WalletSearchRequest,
  TransactionHistoryRequest,
} from '@/types';

export class WalletController {
  private walletService: WalletService;

  constructor() {
    this.walletService = new WalletService();
    this.initializeService();
  }

  private async initializeService(): Promise<void> {
    try {
      await this.walletService.initialize();
    } catch (error) {
      logger.error('Failed to initialize WalletService:', error);
    }
  }

  // Validation schemas
  private walletNetworkSchema = Joi.object({
    address: Joi.string().pattern(/^0x[a-fA-F0-9]{40}$/).required(),
    depth: Joi.number().integer().min(1).max(3).default(2),
  });

  private walletDetailsSchema = Joi.object({
    address: Joi.string().pattern(/^0x[a-fA-F0-9]{40}$/).required(),
  });

  private walletSearchSchema = Joi.object({
    q: Joi.string().min(3).max(100).required(),
    limit: Joi.number().integer().min(1).max(50).default(10),
    offset: Joi.number().integer().min(0).default(0),
  });

  private transactionHistorySchema = Joi.object({
    address: Joi.string().pattern(/^0x[a-fA-F0-9]{40}$/).required(),
    limit: Joi.number().integer().min(1).max(100).default(10),
    offset: Joi.number().integer().min(0).default(0),
    startBlock: Joi.number().integer().min(0).optional(),
    endBlock: Joi.number().integer().min(0).optional(),
  });

  /**
   * GET /api/wallets/network
   * Get wallet network relationships
   */
  public getWalletNetwork = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      // Validate request
      const { error, value } = this.walletNetworkSchema.validate({
        address: req.query.address,
        depth: req.query.depth ? parseInt(req.query.depth as string, 10) : 2,
      });

      if (error) {
        throw new ValidationError(error.details[0].message);
      }

      const request: WalletNetworkRequest = value;

      logger.info(`Getting wallet network for ${request.address} with depth ${request.depth}`);

      // Get network data
      const networkData = await this.walletService.getWalletNetwork(request);

      const response: ApiResponse = {
        success: true,
        data: networkData,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/:address
   * Get wallet details
   */
  public getWalletDetails = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      // Validate request
      const { error, value } = this.walletDetailsSchema.validate({
        address: req.params.address,
      });

      if (error) {
        throw new ValidationError(error.details[0].message);
      }

      logger.info(`Getting wallet details for ${value.address}`);

      // Get wallet details
      const walletDetails = await this.walletService.getWalletDetails(value.address);

      if (!walletDetails) {
        throw new NotFoundError(`Wallet ${value.address} not found`);
      }

      const response: ApiResponse = {
        success: true,
        data: walletDetails,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/search
   * Search wallets by address or label
   */
  public searchWallets = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      // Validate request
      const { error, value } = this.walletSearchSchema.validate({
        q: req.query.q,
        limit: req.query.limit ? parseInt(req.query.limit as string, 10) : 10,
        offset: req.query.offset ? parseInt(req.query.offset as string, 10) : 0,
      });

      if (error) {
        throw new ValidationError(error.details[0].message);
      }

      const request: WalletSearchRequest = {
        query: value.q,
        limit: value.limit,
        offset: value.offset,
      };

      logger.info(`Searching wallets with query: ${request.query}`);

      // Search wallets
      const searchResults = await this.walletService.searchWallets(request);

      const response: ApiResponse = {
        success: true,
        data: searchResults,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/:address/transactions
   * Get wallet transaction history
   */
  public getWalletTransactions = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      // Validate request
      const { error, value } = this.transactionHistorySchema.validate({
        address: req.params.address,
        limit: req.query.limit ? parseInt(req.query.limit as string, 10) : 10,
        offset: req.query.offset ? parseInt(req.query.offset as string, 10) : 0,
        startBlock: req.query.startBlock ? parseInt(req.query.startBlock as string, 10) : undefined,
        endBlock: req.query.endBlock ? parseInt(req.query.endBlock as string, 10) : undefined,
      });

      if (error) {
        throw new ValidationError(error.details[0].message);
      }

      const request: TransactionHistoryRequest = {
        address: value.address,
        limit: value.limit,
        offset: value.offset,
        startBlock: value.startBlock,
        endBlock: value.endBlock,
      };

      logger.info(`Getting transaction history for ${request.address}`);

      // Get transaction history
      const transactions = await this.walletService.getWalletTransactions(request);

      const response: ApiResponse = {
        success: true,
        data: transactions,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * POST /api/wallets/batch
   * Get details for multiple wallets
   */
  public getBatchWalletDetails = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const addresses = req.body.addresses;

      if (!Array.isArray(addresses) || addresses.length === 0) {
        throw new ValidationError('addresses must be a non-empty array');
      }

      if (addresses.length > 20) {
        throw new ValidationError('Maximum 20 addresses allowed per batch request');
      }

      // Validate each address
      for (const address of addresses) {
        if (typeof address !== 'string' || !/^0x[a-fA-F0-9]{40}$/.test(address)) {
          throw new ValidationError(`Invalid address format: ${address}`);
        }
      }

      logger.info(`Getting batch wallet details for ${addresses.length} addresses`);

      // Get details for all wallets
      const walletDetails = await Promise.allSettled(
        addresses.map(address => this.walletService.getWalletDetails(address))
      );

      // Process results
      const results = walletDetails.map((result, index) => {
        if (result.status === 'fulfilled') {
          return {
            address: addresses[index],
            success: true,
            data: result.value,
          };
        } else {
          return {
            address: addresses[index],
            success: false,
            error: result.reason?.message || 'Unknown error',
          };
        }
      });

      const response: ApiResponse = {
        success: true,
        data: results,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/stats
   * Get general statistics about wallets in the system
   */
  public getWalletStats = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      logger.info('Getting wallet statistics');

      // This would typically query the database for statistics
      // For now, return mock data
      const stats = {
        totalWallets: 0,
        totalContracts: 0,
        totalTransactions: 0,
        averageBalance: '0',
        topWalletsByBalance: [],
        topWalletsByTransactions: [],
        recentActivity: [],
      };

      const response: ApiResponse = {
        success: true,
        data: stats,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/popular
   * Get popular wallets from pre-crawled data
   */
  public getPopularWallets = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      logger.info('Getting popular wallets from cache');

      // Try to get from cache first
      const popularWallets = await this.walletService.getPopularWallets();

      const response: ApiResponse = {
        success: true,
        data: popularWallets || [],
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };

  /**
   * GET /api/wallets/market-data
   * Get market data from pre-crawled data
   */
  public getMarketData = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      logger.info('Getting market data from cache');

      // Try to get from cache first
      const marketData = await this.walletService.getMarketData();

      const response: ApiResponse = {
        success: true,
        data: marketData,
        timestamp: new Date().toISOString(),
      };

      res.json(response);

    } catch (error) {
      next(error);
    }
  };
}
