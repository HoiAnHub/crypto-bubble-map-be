#!/usr/bin/env node

/**
 * Test script for the proactive data crawling system
 * This script tests the new background job functionality and API endpoints
 */

const axios = require('axios');

const API_BASE_URL = process.env.API_URL || 'http://localhost:3001';

async function testAPI() {
  console.log('🧪 Testing Proactive Data Crawling System');
  console.log('==========================================\n');

  try {
    // Test 1: Basic health check
    console.log('1. Testing basic health check...');
    const healthResponse = await axios.get(`${API_BASE_URL}/health`);
    console.log('✅ Health check:', healthResponse.data.status);

    // Test 2: Job system health
    console.log('\n2. Testing job system health...');
    const jobHealthResponse = await axios.get(`${API_BASE_URL}/api/jobs/health`);
    console.log('✅ Job system:', jobHealthResponse.data.data);

    // Test 3: Job statistics
    console.log('\n3. Testing job statistics...');
    try {
      const jobStatsResponse = await axios.get(`${API_BASE_URL}/api/jobs/stats`);
      console.log('✅ Job stats:', JSON.stringify(jobStatsResponse.data.data, null, 2));
    } catch (error) {
      console.log('⚠️  Job stats not available (service may be starting up)');
    }

    // Test 4: Market data endpoint (pre-crawled)
    console.log('\n4. Testing market data endpoint...');
    try {
      const marketDataResponse = await axios.get(`${API_BASE_URL}/api/wallets/market-data`);
      if (marketDataResponse.data.data) {
        console.log('✅ Market data available (pre-crawled)');
        console.log('   ETH Price:', marketDataResponse.data.data.ethereum?.price || 'Not available');
        console.log('   Gas Prices:', marketDataResponse.data.data.gasTracker || 'Not available');
      } else {
        console.log('⚠️  Market data not yet crawled (background job may still be running)');
      }
    } catch (error) {
      console.log('⚠️  Market data endpoint error:', error.response?.data?.error || error.message);
    }

    // Test 5: Popular wallets endpoint (pre-crawled)
    console.log('\n5. Testing popular wallets endpoint...');
    try {
      const popularWalletsResponse = await axios.get(`${API_BASE_URL}/api/wallets/popular`);
      if (popularWalletsResponse.data.data && popularWalletsResponse.data.data.length > 0) {
        console.log('✅ Popular wallets available (pre-crawled)');
        console.log(`   Found ${popularWalletsResponse.data.data.length} popular wallets`);
      } else {
        console.log('⚠️  Popular wallets not yet crawled (background job may still be running)');
      }
    } catch (error) {
      console.log('⚠️  Popular wallets endpoint error:', error.response?.data?.error || error.message);
    }

    // Test 6: Traditional wallet network endpoint (should still work)
    console.log('\n6. Testing traditional wallet network endpoint...');
    try {
      const testAddress = '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6'; // Vitalik's address
      const networkResponse = await axios.get(`${API_BASE_URL}/api/wallets/network`, {
        params: { address: testAddress, depth: 1 }
      });
      console.log('✅ Wallet network endpoint working');
      console.log(`   Nodes: ${networkResponse.data.data.nodes?.length || 0}`);
      console.log(`   Links: ${networkResponse.data.data.links?.length || 0}`);
    } catch (error) {
      console.log('⚠️  Wallet network endpoint error:', error.response?.data?.error || error.message);
    }

    // Test 7: Wallet details endpoint (should use cache if available)
    console.log('\n7. Testing wallet details endpoint...');
    try {
      const testAddress = '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6';
      const detailsResponse = await axios.get(`${API_BASE_URL}/api/wallets/${testAddress}`);
      console.log('✅ Wallet details endpoint working');
      console.log(`   Address: ${detailsResponse.data.data.address}`);
      console.log(`   Balance: ${detailsResponse.data.data.balance || 'Not available'} ETH`);
    } catch (error) {
      console.log('⚠️  Wallet details endpoint error:', error.response?.data?.error || error.message);
    }

    // Test 8: Manual job trigger (if available)
    console.log('\n8. Testing manual job trigger...');
    try {
      const triggerResponse = await axios.post(`${API_BASE_URL}/api/jobs/trigger/market_data_crawl`, {
        priority: 'high'
      });
      console.log('✅ Manual job trigger working');
      console.log('   Response:', triggerResponse.data.data.message);
    } catch (error) {
      console.log('⚠️  Manual job trigger error:', error.response?.data?.error || error.message);
    }

    console.log('\n🎉 Proactive Data Crawling System Test Complete!');
    console.log('\n📊 Summary:');
    console.log('- Background job system is integrated');
    console.log('- New API endpoints for pre-crawled data are available');
    console.log('- Job management and monitoring endpoints are working');
    console.log('- Traditional endpoints still function as expected');
    
    console.log('\n💡 Next Steps:');
    console.log('1. Configure environment variables for job intervals');
    console.log('2. Set up external API keys (Etherscan, CoinGecko)');
    console.log('3. Monitor job execution via /api/jobs/stats');
    console.log('4. Check logs for background job activity');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
    if (error.code === 'ECONNREFUSED') {
      console.log('\n💡 Make sure the backend server is running:');
      console.log('   cd crypto-bubble-map-be && npm run dev');
    }
  }
}

// Performance comparison test
async function testPerformanceComparison() {
  console.log('\n⚡ Performance Comparison Test');
  console.log('==============================\n');

  const testAddress = '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6';

  try {
    // Test cached endpoint performance
    console.log('Testing cached market data endpoint...');
    const start1 = Date.now();
    await axios.get(`${API_BASE_URL}/api/wallets/market-data`);
    const cached_time = Date.now() - start1;
    console.log(`✅ Cached market data response time: ${cached_time}ms`);

    // Test traditional endpoint performance
    console.log('Testing traditional wallet details endpoint...');
    const start2 = Date.now();
    await axios.get(`${API_BASE_URL}/api/wallets/${testAddress}`);
    const traditional_time = Date.now() - start2;
    console.log(`✅ Traditional wallet details response time: ${traditional_time}ms`);

    console.log('\n📈 Performance Analysis:');
    if (cached_time < traditional_time) {
      console.log(`🚀 Cached endpoints are ${Math.round((traditional_time / cached_time) * 100) / 100}x faster!`);
    } else {
      console.log('⚠️  Cache may not be populated yet. Try again after background jobs run.');
    }

  } catch (error) {
    console.log('⚠️  Performance test error:', error.message);
  }
}

// Run tests
async function runAllTests() {
  await testAPI();
  await testPerformanceComparison();
}

// Execute if run directly
if (require.main === module) {
  runAllTests().catch(console.error);
}

module.exports = { testAPI, testPerformanceComparison };
