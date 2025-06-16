import { EthereumService } from './EthereumService';
import { Neo4jService } from './Neo4jService';
import { DatabaseService } from './DatabaseService';
import { RedisService } from './RedisService';
import { logger } from '@/utils/logger';
import { config } from '@/config/config';
import {
  WalletNode,
  WalletDetails,
  GraphData,
  Transaction,
  SearchResult,
  TransactionHistoryRequest,
  WalletNetworkRequest,
  WalletSearchRequest,
} from '@/types';

export class WalletService {
  private ethereumService: EthereumService;
  private neo4jService: Neo4jService;
  private databaseService: DatabaseService;
  private redisService: RedisService;

  constructor() {
    this.ethereumService = new EthereumService();
    this.neo4jService = new Neo4jService();
    this.databaseService = new DatabaseService();
    this.redisService = new RedisService();
  }

  public async initialize(): Promise<void> {
    await Promise.all([
      this.neo4jService.connect(),
      this.databaseService.connect(),
      this.redisService.connect(),
    ]);

    // Initialize database tables
    await this.databaseService.initializeTables();
  }

  public async getWalletNetwork(request: WalletNetworkRequest): Promise<GraphData> {
    const { address, depth = 2 } = request;

    // Validate address
    if (!this.ethereumService.isValidAddress(address)) {
      throw new Error('Invalid Ethereum address');
    }

    const normalizedAddress = this.ethereumService.normalizeAddress(address);
    const cacheKey = RedisService.generateKey('wallet_network', normalizedAddress, depth);

    try {
      // Try to get from cache first
      const cached = await this.redisService.get<GraphData>(cacheKey);
      if (cached) {
        logger.debug(`Cache hit for wallet network: ${normalizedAddress}`);
        return cached;
      }

      // Ensure the wallet exists in our system
      await this.ensureWalletExists(normalizedAddress);

      // Get network data from Neo4j
      const networkData = await this.neo4jService.getWalletNetwork(normalizedAddress, Math.min(depth, config.performance.maxNetworkDepth));

      // If no data found, try to build network by analyzing transactions
      if (networkData.nodes.length === 0) {
        logger.info(`No network data found for ${normalizedAddress}, building from transactions`);
        await this.buildWalletNetwork(normalizedAddress, depth);

        // Try again after building
        const newNetworkData = await this.neo4jService.getWalletNetwork(normalizedAddress, depth);

        // Cache the result
        await this.redisService.set(cacheKey, newNetworkData, config.cache.networkDataTtl);

        return newNetworkData;
      }

      // Cache the result
      await this.redisService.set(cacheKey, networkData, config.cache.networkDataTtl);

      return networkData;

    } catch (error) {
      logger.error(`Error getting wallet network for ${normalizedAddress}:`, error);
      throw error;
    }
  }

  public async getWalletDetails(address: string): Promise<WalletNode | null> {
    // Validate address
    if (!this.ethereumService.isValidAddress(address)) {
      throw new Error('Invalid Ethereum address');
    }

    const normalizedAddress = this.ethereumService.normalizeAddress(address);
    const cacheKey = RedisService.generateKey('wallet_details', normalizedAddress);

    try {
      // Try to get from cache first
      const cached = await this.redisService.get<WalletNode>(cacheKey);
      if (cached) {
        logger.debug(`Cache hit for wallet details: ${normalizedAddress}`);
        return cached;
      }

      // Ensure the wallet exists and is up to date
      await this.ensureWalletExists(normalizedAddress);

      // Get details from Neo4j
      const walletDetails = await this.neo4jService.getWalletDetails(normalizedAddress);

      if (walletDetails) {
        // Cache the result
        await this.redisService.set(cacheKey, walletDetails, config.cache.walletDetailsTtl);
      }

      return walletDetails;

    } catch (error) {
      logger.error(`Error getting wallet details for ${normalizedAddress}:`, error);
      throw error;
    }
  }

  public async searchWallets(request: WalletSearchRequest): Promise<SearchResult[]> {
    const { query, limit = 10 } = request;

    if (!query || query.length < 3) {
      throw new Error('Search query must be at least 3 characters long');
    }

    const cacheKey = RedisService.generateKey('wallet_search', query, limit);

    try {
      // Try to get from cache first
      const cached = await this.redisService.get<SearchResult[]>(cacheKey);
      if (cached) {
        logger.debug(`Cache hit for wallet search: ${query}`);
        return cached;
      }

      // Search in Neo4j
      const wallets = await this.neo4jService.searchWallets(query, limit);

      // Transform to search results
      const searchResults: SearchResult[] = wallets.map(wallet => ({
        address: wallet.address,
        label: wallet.label,
        type: wallet.isContract ? 'contract' : 'wallet',
        transactionCount: wallet.transactionCount || 0,
        balance: wallet.balance,
        relevanceScore: this.calculateRelevanceScore(wallet, query),
      }));

      // Sort by relevance score
      searchResults.sort((a, b) => b.relevanceScore - a.relevanceScore);

      // Cache the result
      await this.redisService.set(cacheKey, searchResults, 300); // 5 minutes cache

      return searchResults;

    } catch (error) {
      logger.error(`Error searching wallets with query "${query}":`, error);
      throw error;
    }
  }

  public async getWalletTransactions(request: TransactionHistoryRequest): Promise<Transaction[]> {
    const { address, limit = 10 } = request;

    // Validate address
    if (!this.ethereumService.isValidAddress(address)) {
      throw new Error('Invalid Ethereum address');
    }

    const normalizedAddress = this.ethereumService.normalizeAddress(address);
    const cacheKey = RedisService.generateKey('wallet_transactions', normalizedAddress, limit);

    try {
      // Try to get from cache first
      const cached = await this.redisService.get<Transaction[]>(cacheKey);
      if (cached) {
        logger.debug(`Cache hit for wallet transactions: ${normalizedAddress}`);
        return cached;
      }

      // Get transactions from Ethereum service
      const transactions = await this.ethereumService.getWalletTransactions(
        normalizedAddress,
        Math.min(limit, config.performance.maxTransactionsPerRequest),
        request.startBlock,
        request.endBlock
      );

      // Cache the result
      await this.redisService.set(cacheKey, transactions, config.cache.transactionsTtl);

      return transactions;

    } catch (error) {
      logger.error(`Error getting wallet transactions for ${normalizedAddress}:`, error);
      throw error;
    }
  }

  private async ensureWalletExists(address: string): Promise<void> {
    try {
      // Check if wallet exists in Neo4j
      const existingWallet = await this.neo4jService.getWalletDetails(address);

      if (!existingWallet) {
        logger.info(`Creating new wallet entry for ${address}`);

        // Get wallet details from Ethereum
        const walletDetails = await this.ethereumService.getWalletDetails(address);

        // Create wallet node
        const walletNode: WalletNode = {
          id: '', // Will be set by Neo4j
          address: walletDetails.address,
          balance: walletDetails.balance,
          balanceUSD: walletDetails.balanceUSD,
          transactionCount: walletDetails.transactionCount,
          isContract: walletDetails.isContract,
          contractType: walletDetails.contractType,
          tags: walletDetails.tags,
          riskScore: walletDetails.riskScore,
          firstSeen: new Date(),
          lastActivity: new Date(),
        };

        // Store in Neo4j
        await this.neo4jService.createWalletNode(walletNode);

        // Store in PostgreSQL
        await this.storeWalletInDatabase(walletNode);
      }
    } catch (error) {
      logger.error(`Error ensuring wallet exists for ${address}:`, error);
      throw error;
    }
  }

  private async buildWalletNetwork(address: string, depth: number): Promise<void> {
    try {
      logger.info(`Building wallet network for ${address} with depth ${depth}`);

      // Get recent transactions for the wallet
      const transactions = await this.ethereumService.getWalletTransactions(address, 50);

      // Process transactions and create relationships
      for (const transaction of transactions) {
        // Ensure both wallets exist
        await this.ensureWalletExists(transaction.from);
        if (transaction.to) {
          await this.ensureWalletExists(transaction.to);
        }

        // Create transaction relationship in Neo4j
        await this.neo4jService.createTransactionRelationship(transaction);

        // Store transaction in PostgreSQL
        await this.storeTransactionInDatabase(transaction);
      }

      // If depth > 1, recursively build network for connected wallets
      if (depth > 1) {
        const connectedAddresses = new Set<string>();

        transactions.forEach(tx => {
          if (tx.from !== address) connectedAddresses.add(tx.from);
          if (tx.to && tx.to !== address) connectedAddresses.add(tx.to);
        });

        // Limit the number of connected wallets to process
        const addressesToProcess = Array.from(connectedAddresses).slice(0, 5);

        for (const connectedAddress of addressesToProcess) {
          await this.buildWalletNetwork(connectedAddress, depth - 1);
        }
      }

    } catch (error) {
      logger.error(`Error building wallet network for ${address}:`, error);
      throw error;
    }
  }

  private async storeWalletInDatabase(wallet: WalletNode): Promise<void> {
    try {
      await this.databaseService.query(
        `
        INSERT INTO wallets (
          address, balance, balance_usd, transaction_count, is_contract,
          contract_type, label, tags, risk_score, first_seen, last_activity
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (address) DO UPDATE SET
          balance = EXCLUDED.balance,
          balance_usd = EXCLUDED.balance_usd,
          transaction_count = EXCLUDED.transaction_count,
          is_contract = EXCLUDED.is_contract,
          contract_type = EXCLUDED.contract_type,
          label = EXCLUDED.label,
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
          wallet.contractType || null,
          wallet.label || null,
          wallet.tags || [],
          wallet.riskScore || 0,
          wallet.firstSeen || new Date(),
          wallet.lastActivity || new Date(),
        ]
      );
    } catch (error) {
      logger.error(`Error storing wallet in database:`, error);
      throw error;
    }
  }

  private async storeTransactionInDatabase(transaction: Transaction): Promise<void> {
    try {
      await this.databaseService.query(
        `
        INSERT INTO transactions (
          hash, from_address, to_address, value, value_usd, gas_used,
          gas_price, timestamp, block_number, status, method_id, function_name
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (hash) DO NOTHING
        `,
        [
          transaction.hash,
          transaction.from,
          transaction.to || null,
          transaction.value,
          transaction.valueUSD || 0,
          transaction.gasUsed,
          transaction.gasPrice,
          new Date(transaction.timestamp * 1000),
          transaction.blockNumber,
          transaction.status,
          transaction.methodId || null,
          transaction.functionName || null,
        ]
      );
    } catch (error) {
      logger.error(`Error storing transaction in database:`, error);
      throw error;
    }
  }

  private calculateRelevanceScore(wallet: WalletNode, query: string): number {
    let score = 0;
    const lowerQuery = query.toLowerCase();
    const lowerAddress = wallet.address.toLowerCase();
    const lowerLabel = wallet.label?.toLowerCase() || '';

    // Exact address match
    if (lowerAddress === lowerQuery) {
      score += 100;
    } else if (lowerAddress.includes(lowerQuery)) {
      score += 50;
    }

    // Label match
    if (lowerLabel === lowerQuery) {
      score += 80;
    } else if (lowerLabel.includes(lowerQuery)) {
      score += 40;
    }

    // Boost score based on transaction count
    score += Math.min(wallet.transactionCount || 0, 20);

    // Boost score for contracts
    if (wallet.isContract) {
      score += 10;
    }

    return score;
  }

  /**
   * Get popular wallets from cache
   */
  public async getPopularWallets(): Promise<any> {
    try {
      const cacheKey = 'popular_wallets';
      const cached = await this.redisService.get(cacheKey);

      if (cached) {
        logger.debug('Cache hit for popular wallets');
        return cached;
      }

      logger.debug('No cached popular wallets found');
      return null;

    } catch (error) {
      logger.error('Error getting popular wallets:', error);
      return null;
    }
  }

  /**
   * Get market data from cache
   */
  public async getMarketData(): Promise<any> {
    try {
      const cacheKey = 'market_data:latest';
      const cached = await this.redisService.get(cacheKey);

      if (cached) {
        logger.debug('Cache hit for market data');
        return cached;
      }

      logger.debug('No cached market data found');
      return null;

    } catch (error) {
      logger.error('Error getting market data:', error);
      return null;
    }
  }
}
