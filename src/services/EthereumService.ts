import { ethers } from 'ethers';
import axios from 'axios';
import { config } from '@/config/config';
import { logger } from '@/utils/logger';
import { Transaction, TokenTransfer, WalletDetails, TokenBalance } from '@/types';

export class EthereumService {
  private provider: ethers.JsonRpcProvider | null = null;
  private backupProvider: ethers.JsonRpcProvider | null = null;
  private currentProvider: ethers.JsonRpcProvider | null = null;

  constructor() {
    this.initializeProviders();
  }

  private initializeProviders(): void {
    try {
      if (config.ethereum.rpcUrl) {
        this.provider = new ethers.JsonRpcProvider(config.ethereum.rpcUrl);
        this.currentProvider = this.provider;
        logger.info('Primary Ethereum provider initialized');
      }

      if (config.ethereum.rpcUrlBackup) {
        this.backupProvider = new ethers.JsonRpcProvider(config.ethereum.rpcUrlBackup);
        logger.info('Backup Ethereum provider initialized');
      }

      if (!this.currentProvider) {
        throw new Error('No Ethereum RPC URL configured');
      }
    } catch (error) {
      logger.error('Failed to initialize Ethereum providers:', error);
      throw error;
    }
  }

  private async switchToBackupProvider(): Promise<void> {
    if (this.backupProvider && this.currentProvider !== this.backupProvider) {
      logger.warn('Switching to backup Ethereum provider');
      this.currentProvider = this.backupProvider;
    }
  }

  private async executeWithFallback<T>(operation: () => Promise<T>): Promise<T> {
    try {
      return await operation();
    } catch (error) {
      logger.warn('Primary provider failed, trying backup:', error);
      await this.switchToBackupProvider();
      return await operation();
    }
  }

  public async getWalletBalance(address: string): Promise<string> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      const balance = await this.currentProvider.getBalance(address);
      return ethers.formatEther(balance);
    });
  }

  public async getTransactionCount(address: string): Promise<number> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      return await this.currentProvider.getTransactionCount(address);
    });
  }

  public async isContract(address: string): Promise<boolean> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      const code = await this.currentProvider.getCode(address);
      return code !== '0x';
    });
  }

  public async getTransaction(hash: string): Promise<Transaction | null> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      const [tx, receipt] = await Promise.all([
        this.currentProvider.getTransaction(hash),
        this.currentProvider.getTransactionReceipt(hash)
      ]);

      if (!tx || !receipt) {
        return null;
      }

      const block = await this.currentProvider.getBlock(tx.blockNumber!);

      return {
        hash: tx.hash,
        from: tx.from,
        to: tx.to || '',
        value: ethers.formatEther(tx.value),
        gasUsed: Number(receipt.gasUsed),
        gasPrice: ethers.formatUnits(tx.gasPrice || 0, 'gwei'),
        timestamp: block!.timestamp,
        blockNumber: tx.blockNumber!,
        status: receipt.status || 0,
        methodId: tx.data.slice(0, 10),
      };
    });
  }

  public async getWalletTransactions(
    address: string,
    limit: number = 10,
    startBlock?: number,
    endBlock?: number
  ): Promise<Transaction[]> {
    try {
      // Use Etherscan API for transaction history as it's more efficient
      if (!config.apis.etherscan) {
        logger.warn('Etherscan API key not configured, using limited provider method');
        return await this.getTransactionsFromProvider(address, limit);
      }

      const params = new URLSearchParams({
        module: 'account',
        action: 'txlist',
        address: address,
        startblock: (startBlock || 0).toString(),
        endblock: (endBlock || 99999999).toString(),
        page: '1',
        offset: limit.toString(),
        sort: 'desc',
        apikey: config.apis.etherscan,
      });

      const response = await axios.get(
        `https://api.etherscan.io/api?${params.toString()}`,
        { timeout: 10000 }
      );

      if (response.data.status !== '1') {
        throw new Error(`Etherscan API error: ${response.data.message}`);
      }

      return response.data.result.map((tx: any) => ({
        hash: tx.hash,
        from: tx.from,
        to: tx.to,
        value: ethers.formatEther(tx.value),
        gasUsed: parseInt(tx.gasUsed),
        gasPrice: ethers.formatUnits(tx.gasPrice, 'gwei'),
        timestamp: parseInt(tx.timeStamp),
        blockNumber: parseInt(tx.blockNumber),
        status: parseInt(tx.txreceipt_status || '1'),
        methodId: tx.methodId,
        functionName: tx.functionName,
      }));

    } catch (error) {
      logger.error('Error fetching transactions from Etherscan:', error);
      // Fallback to provider method
      return await this.getTransactionsFromProvider(address, limit);
    }
  }

  private async getTransactionsFromProvider(address: string, limit: number): Promise<Transaction[]> {
    // This is a simplified implementation - in practice, you'd need to scan blocks
    // or use event logs to find transactions for a specific address
    logger.warn('Provider-based transaction fetching is limited - consider using Etherscan API');
    return [];
  }

  public async getTokenBalances(address: string): Promise<TokenBalance[]> {
    try {
      if (!config.apis.etherscan) {
        logger.warn('Etherscan API key not configured, cannot fetch token balances');
        return [];
      }

      const params = new URLSearchParams({
        module: 'account',
        action: 'tokentx',
        address: address,
        page: '1',
        offset: '100',
        sort: 'desc',
        apikey: config.apis.etherscan,
      });

      const response = await axios.get(
        `https://api.etherscan.io/api?${params.toString()}`,
        { timeout: 10000 }
      );

      if (response.data.status !== '1') {
        return [];
      }

      // Group by token address and calculate balances
      const tokenMap = new Map<string, TokenBalance>();

      for (const tx of response.data.result) {
        const tokenAddress = tx.contractAddress;
        const isIncoming = tx.to.toLowerCase() === address.toLowerCase();
        const value = BigInt(tx.value);

        if (!tokenMap.has(tokenAddress)) {
          tokenMap.set(tokenAddress, {
            tokenAddress,
            symbol: tx.tokenSymbol,
            name: tx.tokenName,
            decimals: parseInt(tx.tokenDecimal),
            balance: '0',
            balanceFormatted: '0',
          });
        }

        const tokenBalance = tokenMap.get(tokenAddress)!;
        const currentBalance = BigInt(tokenBalance.balance);
        const newBalance = isIncoming ? currentBalance + value : currentBalance - value;

        tokenBalance.balance = newBalance.toString();
        tokenBalance.balanceFormatted = ethers.formatUnits(newBalance, tokenBalance.decimals);
      }

      return Array.from(tokenMap.values()).filter(token =>
        BigInt(token.balance) > 0
      );

    } catch (error) {
      logger.error('Error fetching token balances:', error);
      return [];
    }
  }

  public async getWalletDetails(address: string): Promise<WalletDetails> {
    try {
      const [balance, transactionCount, isContract, tokenBalances] = await Promise.all([
        this.getWalletBalance(address),
        this.getTransactionCount(address),
        this.isContract(address),
        this.getTokenBalances(address),
      ]);

      // Get ETH price for USD conversion
      const ethPriceUSD = await this.getETHPrice();
      const balanceUSD = parseFloat(balance) * ethPriceUSD;

      return {
        address,
        balance,
        balanceUSD,
        transactionCount,
        isContract,
        tags: [],
        riskScore: 0, // Will be calculated by risk assessment service
        tokenBalances,
      };

    } catch (error) {
      logger.error(`Error getting wallet details for ${address}:`, error);
      throw error;
    }
  }

  public async getETHPrice(): Promise<number> {
    try {
      const response = await axios.get(
        'https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd',
        { timeout: 5000 }
      );

      return response.data.ethereum.usd;
    } catch (error) {
      logger.error('Error fetching ETH price:', error);
      return 0; // Return 0 if price fetch fails
    }
  }

  public async getCurrentBlockNumber(): Promise<number> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      return await this.currentProvider.getBlockNumber();
    });
  }

  public async getBlock(blockNumber: number): Promise<any> {
    return await this.executeWithFallback(async () => {
      if (!this.currentProvider) {
        throw new Error('No Ethereum provider available');
      }

      return await this.currentProvider.getBlock(blockNumber);
    });
  }

  public isValidAddress(address: string): boolean {
    return ethers.isAddress(address);
  }

  public normalizeAddress(address: string): string {
    return ethers.getAddress(address);
  }
}
