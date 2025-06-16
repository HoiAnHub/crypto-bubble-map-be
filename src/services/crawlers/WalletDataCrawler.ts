import { logger } from '@/utils/logger';
import { RedisService } from '../RedisService';
import { DatabaseService } from '../DatabaseService';
import { EthereumService } from '../EthereumService';
import { Neo4jService } from '../Neo4jService';
import { WalletNode, Transaction } from '@/types';
import { config } from '@/config/config';

export interface WalletCrawlResult {
  address: string;
  success: boolean;
  error?: string;
  walletData?: WalletNode;
  transactionCount?: number;
}

export interface PopularWallet {
  address: string;
  transactionCount: number;
  balance: string;
  balanceUSD: number;
  activityScore: number;
  lastActivity: Date;
  tags: string[];
}

export class WalletDataCrawler {
  private redisService: RedisService;
  private databaseService: DatabaseService;
  private ethereumService: EthereumService;
  private neo4jService: Neo4jService;

  constructor(
    redisService: RedisService,
    databaseService: DatabaseService,
    ethereumService: EthereumService,
    neo4jService: Neo4jService
  ) {
    this.redisService = redisService;
    this.databaseService = databaseService;
    this.ethereumService = ethereumService;
    this.neo4jService = neo4jService;
  }

  /**
   * Crawl wallet data for a batch of addresses
   */
  public async crawlWalletBatch(addresses: string[]): Promise<WalletCrawlResult[]> {
    logger.info(`Starting wallet crawl for ${addresses.length} addresses`);

    const results: WalletCrawlResult[] = [];

    for (const address of addresses) {
      try {
        const walletData = await this.crawlSingleWallet(address);
        results.push({
          address,
          success: true,
          walletData,
          transactionCount: walletData.transactionCount,
        });

        // Rate limiting to avoid overwhelming external APIs
        await new Promise(resolve => setTimeout(resolve, 100)); // 100ms delay

      } catch (error) {
        logger.error(`Error crawling wallet ${address}:`, error);
        results.push({
          address,
          success: false,
          error: error instanceof Error ? error.message : 'Unknown error',
        });
      }
    }

    logger.info(`Wallet crawl completed: ${results.filter(r => r.success).length}/${results.length} successful`);
    return results;
  }

  /**
   * Crawl data for a single wallet
   */
  private async crawlSingleWallet(address: string): Promise<WalletNode> {
    const normalizedAddress = this.ethereumService.normalizeAddress(address);

    // Get wallet details from blockchain
    const walletDetails = await this.ethereumService.getWalletDetails(normalizedAddress);

    // Get recent transactions to analyze activity
    const recentTransactions = await this.ethereumService.getWalletTransactions(
      normalizedAddress,
      50 // Get more transactions for better analysis
    );

    // Calculate activity metrics
    const activityMetrics = this.calculateActivityMetrics(recentTransactions);

    // Create wallet node
    const walletNode: WalletNode = {
      id: '', // Will be set by Neo4j
      address: walletDetails.address,
      balance: walletDetails.balance,
      balanceUSD: walletDetails.balanceUSD,
      transactionCount: walletDetails.transactionCount,
      isContract: walletDetails.isContract,
      contractType: walletDetails.contractType,
      tags: this.generateWalletTags(walletDetails, activityMetrics),
      riskScore: this.calculateRiskScore(walletDetails, activityMetrics),
      firstSeen: activityMetrics.firstSeen || new Date(),
      lastActivity: activityMetrics.lastActivity || new Date(),
    };

    // Store in database
    await this.storeWalletInDatabase(walletNode);

    // Store in Neo4j
    await this.neo4jService.createOrUpdateWallet(walletNode);

    // Cache the wallet data
    const cacheKey = RedisService.generateKey('wallet_details', normalizedAddress);
    await this.redisService.set(cacheKey, walletNode, config.cache.walletDetailsTtl);

    // Process transactions for network building
    if (recentTransactions.length > 0) {
      await this.processWalletTransactions(normalizedAddress, recentTransactions);
    }

    return walletNode;
  }

  /**
   * Discover popular wallets based on activity
   */
  public async discoverPopularWallets(): Promise<PopularWallet[]> {
    logger.info('Starting popular wallet discovery');

    try {
      // Get top wallets by transaction count from database
      const topWallets = await this.databaseService.query(
        `
        SELECT address, transaction_count, balance, balance_usd,
               last_activity, tags, created_at
        FROM wallets
        WHERE transaction_count > 100
          AND last_activity > NOW() - INTERVAL '30 days'
        ORDER BY transaction_count DESC, balance_usd DESC
        LIMIT 100
        `,
        []
      );

      const popularWallets: PopularWallet[] = topWallets.rows.map(row => ({
        address: row.address,
        transactionCount: row.transaction_count,
        balance: row.balance,
        balanceUSD: parseFloat(row.balance_usd),
        activityScore: this.calculateActivityScore(row),
        lastActivity: new Date(row.last_activity),
        tags: row.tags || [],
      }));

      // Cache the popular wallets list
      await this.redisService.set('popular_wallets', popularWallets, 21600); // 6 hours TTL

      logger.info(`Discovered ${popularWallets.length} popular wallets`);
      return popularWallets;

    } catch (error) {
      logger.error('Error discovering popular wallets:', error);
      throw error;
    }
  }

  /**
   * Refresh existing wallet data
   */
  public async refreshWalletData(addresses: string[]): Promise<WalletCrawlResult[]> {
    logger.info(`Refreshing data for ${addresses.length} wallets`);

    const results: WalletCrawlResult[] = [];

    for (const address of addresses) {
      try {
        // Check if wallet needs refresh (based on last update time)
        const needsRefresh = await this.checkIfWalletNeedsRefresh(address);

        if (needsRefresh) {
          const walletData = await this.crawlSingleWallet(address);
          results.push({
            address,
            success: true,
            walletData,
            transactionCount: walletData.transactionCount,
          });
        } else {
          logger.debug(`Wallet ${address} doesn't need refresh, skipping`);
        }

        // Rate limiting
        await new Promise(resolve => setTimeout(resolve, 200)); // 200ms delay

      } catch (error) {
        logger.error(`Error refreshing wallet ${address}:`, error);
        results.push({
          address,
          success: false,
          error: error instanceof Error ? error.message : 'Unknown error',
        });
      }
    }

    logger.info(`Wallet refresh completed: ${results.filter(r => r.success).length} wallets refreshed`);
    return results;
  }

  /**
   * Calculate activity metrics from transactions
   */
  private calculateActivityMetrics(transactions: Transaction[]): {
    firstSeen: Date | null;
    lastActivity: Date | null;
    avgTransactionValue: number;
    uniqueCounterparties: number;
  } {
    if (transactions.length === 0) {
      return {
        firstSeen: null,
        lastActivity: null,
        avgTransactionValue: 0,
        uniqueCounterparties: 0,
      };
    }

    const sortedTxs = transactions.sort((a, b) => a.timestamp - b.timestamp);
    const firstSeen = new Date(sortedTxs[0].timestamp * 1000);
    const lastActivity = new Date(sortedTxs[sortedTxs.length - 1].timestamp * 1000);

    const totalValue = transactions.reduce((sum, tx) => sum + parseFloat(tx.value), 0);
    const avgTransactionValue = totalValue / transactions.length;

    const counterparties = new Set<string>();
    transactions.forEach(tx => {
      counterparties.add(tx.from);
      counterparties.add(tx.to);
    });

    return {
      firstSeen,
      lastActivity,
      avgTransactionValue,
      uniqueCounterparties: counterparties.size,
    };
  }

  /**
   * Generate wallet tags based on analysis
   */
  private generateWalletTags(walletDetails: any, activityMetrics: any): string[] {
    const tags: string[] = [];

    if (walletDetails.isContract) {
      tags.push('contract');
      if (walletDetails.contractType) {
        tags.push(walletDetails.contractType);
      }
    }

    if (parseFloat(walletDetails.balance) > 100) {
      tags.push('high-balance');
    }

    if (walletDetails.transactionCount > 1000) {
      tags.push('high-activity');
    }

    if (activityMetrics.uniqueCounterparties > 50) {
      tags.push('hub');
    }

    if (activityMetrics.avgTransactionValue > 10) {
      tags.push('high-value-txs');
    }

    return tags;
  }

  /**
   * Calculate risk score for a wallet
   */
  private calculateRiskScore(walletDetails: any, activityMetrics: any): number {
    let riskScore = 0;

    // High transaction count might indicate automated activity
    if (walletDetails.transactionCount > 10000) {
      riskScore += 20;
    }

    // Very new wallets with high activity
    if (activityMetrics.firstSeen &&
        Date.now() - activityMetrics.firstSeen.getTime() < 7 * 24 * 60 * 60 * 1000 && // Less than 7 days old
        walletDetails.transactionCount > 100) {
      riskScore += 30;
    }

    // Contracts have different risk profiles
    if (walletDetails.isContract) {
      riskScore -= 10; // Generally lower risk
    }

    return Math.max(0, Math.min(100, riskScore));
  }

  /**
   * Calculate activity score for popular wallet ranking
   */
  private calculateActivityScore(walletRow: any): number {
    const txWeight = Math.log(walletRow.transaction_count + 1) * 10;
    const balanceWeight = Math.log(parseFloat(walletRow.balance_usd) + 1) * 5;
    const recentActivityWeight = this.getRecentActivityWeight(walletRow.last_activity);

    return txWeight + balanceWeight + recentActivityWeight;
  }

  /**
   * Get weight based on how recent the last activity was
   */
  private getRecentActivityWeight(lastActivity: Date): number {
    const daysSinceActivity = (Date.now() - new Date(lastActivity).getTime()) / (24 * 60 * 60 * 1000);

    if (daysSinceActivity < 1) return 50;
    if (daysSinceActivity < 7) return 30;
    if (daysSinceActivity < 30) return 10;
    return 0;
  }

  /**
   * Check if wallet needs refresh based on last update time
   */
  private async checkIfWalletNeedsRefresh(address: string): Promise<boolean> {
    try {
      const result = await this.databaseService.query(
        'SELECT updated_at FROM wallets WHERE address = $1',
        [address]
      );

      if (result.rows.length === 0) {
        return true; // Wallet doesn't exist, needs crawling
      }

      const lastUpdate = new Date(result.rows[0].updated_at);
      const hoursSinceUpdate = (Date.now() - lastUpdate.getTime()) / (60 * 60 * 1000);

      // Refresh if older than 2 hours
      return hoursSinceUpdate > 2;

    } catch (error) {
      logger.error(`Error checking wallet refresh status for ${address}:`, error);
      return true; // Default to refresh on error
    }
  }

  /**
   * Store wallet in database
   */
  private async storeWalletInDatabase(wallet: WalletNode): Promise<void> {
    try {
      await this.databaseService.query(
        `
        INSERT INTO wallets (
          address, balance, balance_usd, transaction_count, is_contract,
          contract_type, tags, risk_score, first_seen, last_activity
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (address) DO UPDATE SET
          balance = EXCLUDED.balance,
          balance_usd = EXCLUDED.balance_usd,
          transaction_count = EXCLUDED.transaction_count,
          is_contract = EXCLUDED.is_contract,
          contract_type = EXCLUDED.contract_type,
          tags = EXCLUDED.tags,
          risk_score = EXCLUDED.risk_score,
          last_activity = EXCLUDED.last_activity,
          updated_at = NOW()
        `,
        [
          wallet.address,
          wallet.balance || '0',
          wallet.balanceUSD || 0,
          wallet.transactionCount || 0,
          wallet.isContract || false,
          wallet.contractType,
          wallet.tags || [],
          wallet.riskScore || 0,
          wallet.firstSeen,
          wallet.lastActivity,
        ]
      );

    } catch (error) {
      logger.error(`Error storing wallet ${wallet.address} in database:`, error);
      throw error;
    }
  }

  /**
   * Process wallet transactions for network building
   */
  private async processWalletTransactions(address: string, transactions: Transaction[]): Promise<void> {
    // This would build the transaction network in Neo4j
    // For now, just cache the transactions
    const cacheKey = RedisService.generateKey('wallet_transactions', address);
    await this.redisService.set(cacheKey, transactions, config.cache.transactionsTtl);
  }
}
