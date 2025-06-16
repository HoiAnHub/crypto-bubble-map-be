# Crypto Bubble Map Backend

A Node.js backend service that provides Ethereum blockchain data for the crypto-bubble-map frontend visualization tool.

## Features

- **Ethereum Blockchain Integration**: Real-time data fetching from Ethereum mainnet
- **Wallet Network Analysis**: Discover relationships between wallet addresses
- **Transaction History**: Detailed transaction data and analysis
- **Graph Database**: Neo4j integration for complex relationship queries
- **Caching Layer**: Redis caching for optimal performance
- **RESTful API**: Clean API endpoints matching frontend requirements

## Technology Stack

- **Runtime**: Node.js with TypeScript
- **Framework**: Express.js
- **Blockchain**: Ethers.js for Ethereum integration
- **Databases**: PostgreSQL (primary data) + Neo4j (graph relationships)
- **Caching**: Redis
- **API Documentation**: Swagger/OpenAPI

## API Endpoints

- `GET /api/wallets/network?address={address}&depth={depth}` - Get wallet network relationships
- `GET /api/wallets/{address}` - Get detailed wallet information
- `GET /api/wallets/search?q={query}` - Search wallets by address/label
- `GET /api/wallets/{address}/transactions?limit={limit}` - Get transaction history

## Getting Started

### Prerequisites

- Node.js (v18 or later)
- PostgreSQL database
- Neo4j database
- Redis server
- Ethereum RPC endpoint (Infura, Alchemy, etc.)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/HoiAnHub/crypto-bubble-map-be.git
   cd crypto-bubble-map-be
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations:
   ```bash
   npm run migrate
   ```

5. Start the development server:
   ```bash
   npm run dev
   ```

The server will start on `http://localhost:3001`

## Environment Variables

See `.env.example` for required environment variables including:
- Database connection strings
- Ethereum RPC endpoints
- Redis configuration
- API keys and secrets

## Development

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run test` - Run tests
- `npm run lint` - Run ESLint

## License

ISC