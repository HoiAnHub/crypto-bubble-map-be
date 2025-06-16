import neo4j, { Driver, Session, Result } from 'neo4j-driver';
import { config } from '@/config/config';
import { logger } from '@/utils/logger';
import { WalletNode, TransactionLink, GraphData, Transaction } from '@/types';

export class Neo4jService {
  private driver: Driver | null = null;
  private session: Session | null = null;
  private isConnected = false;

  constructor() {
    // Initialize in the connect method to avoid issues with SSR
  }

  public async connect(): Promise<void> {
    if (this.isConnected && this.driver) {
      return;
    }

    try {
      this.driver = neo4j.driver(
        config.neo4j.uri,
        neo4j.auth.basic(config.neo4j.user, config.neo4j.password),
        {
          maxConnectionLifetime: 3 * 60 * 60 * 1000, // 3 hours
          maxConnectionPoolSize: 50,
          connectionAcquisitionTimeout: 2 * 60 * 1000, // 2 minutes
          disableLosslessIntegers: true,
        }
      );

      // Test the connection
      const session = this.driver.session();
      await session.run('RETURN 1');
      await session.close();

      this.isConnected = true;
      logger.info('✅ Neo4j connected successfully');

    } catch (error) {
      logger.error('Failed to connect to Neo4j:', error);
      this.isConnected = false;
      throw error;
    }
  }

  public async disconnect(): Promise<void> {
    try {
      if (this.session) {
        await this.session.close();
        this.session = null;
      }
      if (this.driver) {
        await this.driver.close();
        this.driver = null;
      }
      this.isConnected = false;
      logger.info('Neo4j disconnected');
    } catch (error) {
      logger.error('Error disconnecting from Neo4j:', error);
    }
  }

  private async getSession(): Promise<Session> {
    if (!this.isConnected || !this.driver) {
      await this.connect();
    }

    if (!this.driver) {
      throw new Error('Neo4j driver not available');
    }

    return this.driver.session();
  }

  public async createWalletNode(wallet: WalletNode): Promise<void> {
    const session = await this.getSession();
    
    try {
      await session.run(
        `
        MERGE (w:Wallet {address: $address})
        SET w.balance = $balance,
            w.balanceUSD = $balanceUSD,
            w.transactionCount = $transactionCount,
            w.isContract = $isContract,
            w.contractType = $contractType,
            w.label = $label,
            w.tags = $tags,
            w.riskScore = $riskScore,
            w.firstSeen = datetime($firstSeen),
            w.lastActivity = datetime($lastActivity),
            w.updatedAt = datetime()
        `,
        {
          address: wallet.address,
          balance: wallet.balance || '0',
          balanceUSD: wallet.balanceUSD || 0,
          transactionCount: wallet.transactionCount || 0,
          isContract: wallet.isContract || false,
          contractType: wallet.contractType || null,
          label: wallet.label || null,
          tags: wallet.tags || [],
          riskScore: wallet.riskScore || 0,
          firstSeen: wallet.firstSeen?.toISOString() || null,
          lastActivity: wallet.lastActivity?.toISOString() || null,
        }
      );
    } finally {
      await session.close();
    }
  }

  public async createTransactionRelationship(transaction: Transaction): Promise<void> {
    const session = await this.getSession();
    
    try {
      await session.run(
        `
        MERGE (from:Wallet {address: $fromAddress})
        MERGE (to:Wallet {address: $toAddress})
        MERGE (from)-[r:TRANSFERS {hash: $hash}]->(to)
        SET r.value = $value,
            r.valueUSD = $valueUSD,
            r.gasUsed = $gasUsed,
            r.gasPrice = $gasPrice,
            r.timestamp = datetime($timestamp),
            r.blockNumber = $blockNumber,
            r.status = $status,
            r.methodId = $methodId,
            r.functionName = $functionName
        `,
        {
          fromAddress: transaction.from,
          toAddress: transaction.to,
          hash: transaction.hash,
          value: transaction.value,
          valueUSD: transaction.valueUSD || 0,
          gasUsed: transaction.gasUsed,
          gasPrice: transaction.gasPrice,
          timestamp: new Date(transaction.timestamp * 1000).toISOString(),
          blockNumber: transaction.blockNumber,
          status: transaction.status,
          methodId: transaction.methodId || null,
          functionName: transaction.functionName || null,
        }
      );
    } finally {
      await session.close();
    }
  }

  public async getWalletNetwork(address: string, depth: number = 2): Promise<GraphData> {
    const session = await this.getSession();
    
    try {
      const result = await session.run(
        `
        MATCH path = (source:Wallet {address: $address})-[r:TRANSFERS*1..${depth}]-(target:Wallet)
        WITH COLLECT(path) as paths
        UNWIND paths as path
        WITH DISTINCT nodes(path) as pathNodes, relationships(path) as pathRels
        UNWIND pathNodes as node
        WITH COLLECT(DISTINCT node) as allNodes, pathRels
        UNWIND pathRels as rel
        WITH allNodes, COLLECT(DISTINCT rel) as allRels
        RETURN allNodes, allRels
        `,
        { address }
      );

      if (result.records.length === 0) {
        return { nodes: [], links: [] };
      }

      const record = result.records[0];
      const nodes = record.get('allNodes') || [];
      const relationships = record.get('allRels') || [];

      // Transform Neo4j nodes to our format
      const walletNodes: WalletNode[] = nodes.map((node: any) => ({
        id: node.identity.toString(),
        address: node.properties.address,
        label: node.properties.label || undefined,
        balance: node.properties.balance || '0',
        balanceUSD: node.properties.balanceUSD || 0,
        transactionCount: node.properties.transactionCount || 0,
        tags: node.properties.tags || [],
        isContract: node.properties.isContract || false,
        contractType: node.properties.contractType || undefined,
        riskScore: node.properties.riskScore || 0,
        size: this.calculateNodeSize(node.properties.transactionCount || 0),
        color: this.assignNodeColor(node.properties),
      }));

      // Transform Neo4j relationships to our format
      const transactionLinks: TransactionLink[] = relationships.map((rel: any) => ({
        source: rel.startNodeId.toString(),
        target: rel.endNodeId.toString(),
        value: parseFloat(rel.properties.value || '0'),
        valueUSD: rel.properties.valueUSD || 0,
        type: 'transfer',
        timestamp: rel.properties.timestamp ? new Date(rel.properties.timestamp).getTime() / 1000 : Date.now() / 1000,
        transactionHash: rel.properties.hash,
        gasUsed: rel.properties.gasUsed || 0,
        gasPrice: rel.properties.gasPrice || '0',
        blockNumber: rel.properties.blockNumber || 0,
      }));

      return { nodes: walletNodes, links: transactionLinks };

    } finally {
      await session.close();
    }
  }

  public async getWalletDetails(address: string): Promise<WalletNode | null> {
    const session = await this.getSession();
    
    try {
      const result = await session.run(
        `
        MATCH (w:Wallet {address: $address})
        RETURN w
        `,
        { address }
      );

      if (result.records.length === 0) {
        return null;
      }

      const wallet = result.records[0].get('w');

      return {
        id: wallet.identity.toString(),
        address: wallet.properties.address,
        label: wallet.properties.label || undefined,
        balance: wallet.properties.balance || '0',
        balanceUSD: wallet.properties.balanceUSD || 0,
        transactionCount: wallet.properties.transactionCount || 0,
        tags: wallet.properties.tags || [],
        isContract: wallet.properties.isContract || false,
        contractType: wallet.properties.contractType || undefined,
        riskScore: wallet.properties.riskScore || 0,
        firstSeen: wallet.properties.firstSeen ? new Date(wallet.properties.firstSeen) : undefined,
        lastActivity: wallet.properties.lastActivity ? new Date(wallet.properties.lastActivity) : undefined,
      };

    } finally {
      await session.close();
    }
  }

  public async searchWallets(query: string, limit: number = 10): Promise<WalletNode[]> {
    const session = await this.getSession();
    
    try {
      const result = await session.run(
        `
        MATCH (w:Wallet)
        WHERE w.address CONTAINS $query 
           OR w.label CONTAINS $query
        RETURN w
        ORDER BY w.transactionCount DESC
        LIMIT $limit
        `,
        { query: query.toLowerCase(), limit }
      );

      return result.records.map(record => {
        const wallet = record.get('w');
        return {
          id: wallet.identity.toString(),
          address: wallet.properties.address,
          label: wallet.properties.label || undefined,
          balance: wallet.properties.balance || '0',
          balanceUSD: wallet.properties.balanceUSD || 0,
          transactionCount: wallet.properties.transactionCount || 0,
          tags: wallet.properties.tags || [],
          isContract: wallet.properties.isContract || false,
          contractType: wallet.properties.contractType || undefined,
          riskScore: wallet.properties.riskScore || 0,
        };
      });

    } finally {
      await session.close();
    }
  }

  public async getTransactionHistory(address: string, limit: number = 10): Promise<TransactionLink[]> {
    const session = await this.getSession();
    
    try {
      const result = await session.run(
        `
        MATCH (w:Wallet {address: $address})-[r:TRANSFERS]-(other:Wallet)
        RETURN r, other
        ORDER BY r.timestamp DESC
        LIMIT $limit
        `,
        { address, limit }
      );

      return result.records.map(record => {
        const rel = record.get('r');
        const other = record.get('other');
        
        return {
          source: rel.startNodeId.toString(),
          target: rel.endNodeId.toString(),
          value: parseFloat(rel.properties.value || '0'),
          valueUSD: rel.properties.valueUSD || 0,
          type: 'transfer',
          timestamp: rel.properties.timestamp ? new Date(rel.properties.timestamp).getTime() / 1000 : Date.now() / 1000,
          transactionHash: rel.properties.hash,
          gasUsed: rel.properties.gasUsed || 0,
          gasPrice: rel.properties.gasPrice || '0',
          blockNumber: rel.properties.blockNumber || 0,
        };
      });

    } finally {
      await session.close();
    }
  }

  private calculateNodeSize(txCount: number): number {
    const minSize = 5;
    const maxSize = 30;
    
    if (txCount === 0) return minSize;
    const logSize = Math.log10(txCount + 1) * 5;
    
    return Math.max(minSize, Math.min(maxSize, logSize));
  }

  private assignNodeColor(properties: any): string {
    if (properties.isContract) {
      return '#8B5CF6'; // purple for contracts
    }
    
    if (properties.label) {
      return '#10B981'; // green for labeled wallets
    }
    
    if (properties.balanceUSD && properties.balanceUSD > 10000) {
      return '#F59E0B'; // amber for high-value wallets
    }
    
    return '#3B82F6'; // blue default
  }

  public async healthCheck(): Promise<boolean> {
    try {
      const session = await this.getSession();
      await session.run('RETURN 1');
      await session.close();
      return true;
    } catch (error) {
      logger.error('Neo4j health check failed:', error);
      return false;
    }
  }
}
