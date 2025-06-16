import { Router } from 'express';
import { WalletController } from '@/controllers/WalletController';

const router = Router();
const walletController = new WalletController();

/**
 * @route GET /api/wallets/network
 * @desc Get wallet network relationships
 * @query address - Ethereum wallet address (required)
 * @query depth - Network depth (optional, default: 2, max: 3)
 */
router.get('/network', walletController.getWalletNetwork);

/**
 * @route GET /api/wallets/search
 * @desc Search wallets by address or label
 * @query q - Search query (required, min 3 chars)
 * @query limit - Number of results (optional, default: 10, max: 50)
 * @query offset - Pagination offset (optional, default: 0)
 */
router.get('/search', walletController.searchWallets);

/**
 * @route GET /api/wallets/stats
 * @desc Get general wallet statistics
 */
router.get('/stats', walletController.getWalletStats);

/**
 * @route POST /api/wallets/batch
 * @desc Get details for multiple wallets
 * @body addresses - Array of wallet addresses (max 20)
 */
router.post('/batch', walletController.getBatchWalletDetails);

/**
 * @route GET /api/wallets/:address
 * @desc Get wallet details
 * @param address - Ethereum wallet address
 */
router.get('/:address', walletController.getWalletDetails);

/**
 * @route GET /api/wallets/:address/transactions
 * @desc Get wallet transaction history
 * @param address - Ethereum wallet address
 * @query limit - Number of transactions (optional, default: 10, max: 100)
 * @query offset - Pagination offset (optional, default: 0)
 * @query startBlock - Start block number (optional)
 * @query endBlock - End block number (optional)
 */
router.get('/:address/transactions', walletController.getWalletTransactions);

export { router as walletRoutes };
