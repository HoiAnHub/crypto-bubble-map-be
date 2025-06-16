import axios from 'axios';
import { logger } from '@/utils/logger';
import { RedisService } from '../RedisService';
import { DatabaseService } from '../DatabaseService';
import { config } from '@/config/config';

export interface MarketData {
  ethereum: {
    price: number;
    marketCap: number;
    volume24h: number;
    priceChange24h: number;
    priceChangePercentage24h: number;
    lastUpdated: Date;
  };
  gasTracker: {
    slow: number;
    standard: number;
    fast: number;
    instant: number;
    lastUpdated: Date;
  };
  networkStats: {
    blockNumber: number;
    blockTime: number;
    difficulty: string;
    hashRate: string;
    lastUpdated: Date;
  };
}

export interface TokenPrice {
  address: string;
  symbol: string;
  name: string;
  price: number;
  priceChange24h: number;
  volume24h: number;
  marketCap: number;
  lastUpdated: Date;
}

export class MarketDataCrawler {
  private redisService: RedisService;
  private databaseService: DatabaseService;
  private rateLimitMap: Map<string, number> = new Map();

  constructor(redisService: RedisService, databaseService: DatabaseService) {
    this.redisService = redisService;
    this.databaseService = databaseService;
  }

  /**
   * Crawl and cache market data
   */
  public async crawlMarketData(): Promise<MarketData> {
    logger.info('Starting market data crawl');

    try {
      const [ethereumData, gasData, networkData] = await Promise.allSettled([
        this.fetchEthereumPrice(),
        this.fetchGasTracker(),
        this.fetchNetworkStats(),
      ]);

      const marketData: MarketData = {
        ethereum: ethereumData.status === 'fulfilled' ? ethereumData.value : this.getDefaultEthereumData(),
        gasTracker: gasData.status === 'fulfilled' ? gasData.value : this.getDefaultGasData(),
        networkStats: networkData.status === 'fulfilled' ? networkData.value : this.getDefaultNetworkData(),
      };

      // Cache the market data
      await this.cacheMarketData(marketData);

      // Store in database for historical tracking
      await this.storeMarketDataHistory(marketData);

      logger.info('Market data crawl completed successfully');
      return marketData;

    } catch (error) {
      logger.error('Error during market data crawl:', error);
      throw error;
    }
  }

  /**
   * Fetch Ethereum price and market data from CoinGecko
   */
  private async fetchEthereumPrice(): Promise<MarketData['ethereum']> {
    try {
      // Apply rate limiting
      await this.applyRateLimit('coingecko', 1200);

      const response = await axios.get(
        'https://api.coingecko.com/api/v3/simple/price',
        {
          params: {
            ids: 'ethereum',
            vs_currencies: 'usd',
            include_market_cap: true,
            include_24hr_vol: true,
            include_24hr_change: true,
          },
          timeout: 15000,
          headers: {
            'User-Agent': 'crypto-bubble-map-backend/1.0.0',
            'Accept': 'application/json',
          },
        }
      );

      const ethData = response.data.ethereum;

      return {
        price: ethData.usd,
        marketCap: ethData.usd_market_cap,
        volume24h: ethData.usd_24h_vol,
        priceChange24h: ethData.usd_24h_change,
        priceChangePercentage24h: ethData.usd_24h_change,
        lastUpdated: new Date(),
      };

    } catch (error) {
      logger.error('Error fetching Ethereum price:', error);
      throw error;
    }
  }

  /**
   * Fetch gas tracker data from Etherscan
   */
  private async fetchGasTracker(): Promise<MarketData['gasTracker']> {
    try {
      if (!config.apis.etherscan) {
        throw new Error('Etherscan API key not configured');
      }

      // Apply rate limiting
      await this.applyRateLimit('etherscan', 200);

      const response = await axios.get(
        'https://api.etherscan.io/api',
        {
          params: {
            module: 'gastracker',
            action: 'gasoracle',
            apikey: config.apis.etherscan,
          },
          timeout: 15000,
          headers: {
            'User-Agent': 'crypto-bubble-map-backend/1.0.0',
            'Accept': 'application/json',
          },
        }
      );

      if (response.data.status !== '1') {
        throw new Error(`Etherscan API error: ${response.data.message}`);
      }

      const gasData = response.data.result;

      return {
        slow: parseInt(gasData.SafeGasPrice),
        standard: parseInt(gasData.ProposeGasPrice),
        fast: parseInt(gasData.FastGasPrice),
        instant: parseInt(gasData.FastGasPrice) + 5, // Estimate instant as fast + 5
        lastUpdated: new Date(),
      };

    } catch (error) {
      logger.error('Error fetching gas tracker data:', error);
      throw error;
    }
  }

  /**
   * Fetch network statistics
   */
  private async fetchNetworkStats(): Promise<MarketData['networkStats']> {
    try {
      // This would typically fetch from your Ethereum service
      // For now, return mock data
      return {
        blockNumber: 0,
        blockTime: 12,
        difficulty: '0',
        hashRate: '0',
        lastUpdated: new Date(),
      };

    } catch (error) {
      logger.error('Error fetching network stats:', error);
      throw error;
    }
  }

  /**
   * Cache market data in Redis
   */
  private async cacheMarketData(marketData: MarketData): Promise<void> {
    const cacheKey = 'market_data:latest';
    await this.redisService.set(cacheKey, marketData, 300); // 5 minutes TTL

    // Cache individual components with different TTLs
    await this.redisService.set('market_data:ethereum', marketData.ethereum, 300);
    await this.redisService.set('market_data:gas_tracker', marketData.gasTracker, 600); // 10 minutes
    await this.redisService.set('market_data:network_stats', marketData.networkStats, 300);
  }

  /**
   * Store market data history in database
   */
  private async storeMarketDataHistory(marketData: MarketData): Promise<void> {
    try {
      await this.databaseService.query(
        `
        INSERT INTO market_data_history (
          eth_price, eth_market_cap, eth_volume_24h, eth_price_change_24h,
          gas_slow, gas_standard, gas_fast, gas_instant,
          block_number, block_time, difficulty, hash_rate,
          created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
        `,
        [
          marketData.ethereum.price,
          Math.round(marketData.ethereum.marketCap), // Convert to integer for BIGINT column
          Math.round(marketData.ethereum.volume24h), // Convert to integer for BIGINT column
          marketData.ethereum.priceChange24h,
          marketData.gasTracker.slow,
          marketData.gasTracker.standard,
          marketData.gasTracker.fast,
          marketData.gasTracker.instant,
          marketData.networkStats.blockNumber,
          marketData.networkStats.blockTime,
          marketData.networkStats.difficulty,
          marketData.networkStats.hashRate,
        ]
      );

    } catch (error) {
      logger.error('Error storing market data history:', error);
      // Don't throw - this is not critical for the crawling process
    }
  }

  /**
   * Get cached market data
   */
  public async getCachedMarketData(): Promise<MarketData | null> {
    try {
      const cacheKey = 'market_data:latest';
      return await this.redisService.get<MarketData>(cacheKey);
    } catch (error) {
      logger.error('Error getting cached market data:', error);
      return null;
    }
  }

  /**
   * Crawl specific token prices
   */
  public async crawlTokenPrices(tokenAddresses: string[]): Promise<TokenPrice[]> {
    logger.info(`Crawling prices for ${tokenAddresses.length} tokens`);

    const tokenPrices: TokenPrice[] = [];

    // Process tokens in batches to avoid rate limits
    const batchSize = 10;
    for (let i = 0; i < tokenAddresses.length; i += batchSize) {
      const batch = tokenAddresses.slice(i, i + batchSize);

      try {
        const batchPrices = await this.fetchTokenPricesBatch(batch);
        tokenPrices.push(...batchPrices);

        // Cache individual token prices
        for (const tokenPrice of batchPrices) {
          const cacheKey = `token_price:${tokenPrice.address}`;
          await this.redisService.set(cacheKey, tokenPrice, 300); // 5 minutes TTL
        }

        // Rate limiting delay
        if (i + batchSize < tokenAddresses.length) {
          await new Promise(resolve => setTimeout(resolve, 1000)); // 1 second delay
        }

      } catch (error) {
        logger.error(`Error fetching token prices for batch ${i}-${i + batchSize}:`, error);
      }
    }

    logger.info(`Successfully crawled prices for ${tokenPrices.length} tokens`);
    return tokenPrices;
  }

  /**
   * Fetch token prices for a batch of addresses
   */
  private async fetchTokenPricesBatch(addresses: string[]): Promise<TokenPrice[]> {
    // This would integrate with a token price API like CoinGecko or DeFiPulse
    // For now, return mock data
    return addresses.map(address => ({
      address,
      symbol: 'UNKNOWN',
      name: 'Unknown Token',
      price: 0,
      priceChange24h: 0,
      volume24h: 0,
      marketCap: 0,
      lastUpdated: new Date(),
    }));
  }

  // Default data methods
  private getDefaultEthereumData(): MarketData['ethereum'] {
    return {
      price: 0,
      marketCap: 0,
      volume24h: 0,
      priceChange24h: 0,
      priceChangePercentage24h: 0,
      lastUpdated: new Date(),
    };
  }

  private getDefaultGasData(): MarketData['gasTracker'] {
    return {
      slow: 20,
      standard: 25,
      fast: 30,
      instant: 35,
      lastUpdated: new Date(),
    };
  }

  private getDefaultNetworkData(): MarketData['networkStats'] {
    return {
      blockNumber: 0,
      blockTime: 12,
      difficulty: '0',
      hashRate: '0',
      lastUpdated: new Date(),
    };
  }

  /**
   * Apply rate limiting for external API calls
   */
  private async applyRateLimit(key: string, delay: number): Promise<void> {
    const now = Date.now();
    const lastCall = this.rateLimitMap.get(key) || 0;
    const timeSinceLastCall = now - lastCall;

    if (timeSinceLastCall < delay) {
      const waitTime = delay - timeSinceLastCall;
      logger.debug(`Rate limiting ${key}: waiting ${waitTime}ms`);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }

    this.rateLimitMap.set(key, Date.now());
  }
}
