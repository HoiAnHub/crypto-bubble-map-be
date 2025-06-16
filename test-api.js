const axios = require('axios');

const API_BASE_URL = 'http://localhost:3001';

async function testAPI() {
  console.log('🧪 Testing Crypto Bubble Map Backend API...\n');

  try {
    // Test 1: Health check
    console.log('1. Testing health endpoint...');
    const healthResponse = await axios.get(`${API_BASE_URL}/health`);
    console.log('✅ Health check:', healthResponse.data.status);

    // Test 2: Detailed health check
    console.log('\n2. Testing detailed health endpoint...');
    try {
      const detailedHealthResponse = await axios.get(`${API_BASE_URL}/health/detailed`);
      console.log('✅ Detailed health check:', detailedHealthResponse.data.status);
      console.log('   Services:', Object.keys(detailedHealthResponse.data.services || {}));
    } catch (error) {
      console.log('⚠️  Detailed health check failed (expected if databases not running)');
    }

    // Test 3: Root endpoint
    console.log('\n3. Testing root endpoint...');
    const rootResponse = await axios.get(`${API_BASE_URL}/`);
    console.log('✅ Root endpoint:', rootResponse.data.message);

    // Test 4: Wallet details (using a well-known Ethereum address)
    console.log('\n4. Testing wallet details endpoint...');
    const testAddress = '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045'; // Vitalik's address
    try {
      const walletResponse = await axios.get(`${API_BASE_URL}/api/wallets/${testAddress}`);
      console.log('✅ Wallet details retrieved for:', testAddress);
      console.log('   Address:', walletResponse.data.data?.address);
    } catch (error) {
      console.log('⚠️  Wallet details failed (expected without blockchain connection):', error.response?.status);
    }

    // Test 5: Wallet network
    console.log('\n5. Testing wallet network endpoint...');
    try {
      const networkResponse = await axios.get(`${API_BASE_URL}/api/wallets/network?address=${testAddress}&depth=2`);
      console.log('✅ Wallet network retrieved');
      console.log('   Nodes:', networkResponse.data.data?.nodes?.length || 0);
      console.log('   Links:', networkResponse.data.data?.links?.length || 0);
    } catch (error) {
      console.log('⚠️  Wallet network failed (expected without databases):', error.response?.status);
    }

    // Test 6: Wallet search
    console.log('\n6. Testing wallet search endpoint...');
    try {
      const searchResponse = await axios.get(`${API_BASE_URL}/api/wallets/search?q=0xd8d`);
      console.log('✅ Wallet search completed');
      console.log('   Results:', searchResponse.data.data?.length || 0);
    } catch (error) {
      console.log('⚠️  Wallet search failed (expected without databases):', error.response?.status);
    }

    // Test 7: Wallet transactions
    console.log('\n7. Testing wallet transactions endpoint...');
    try {
      const transactionsResponse = await axios.get(`${API_BASE_URL}/api/wallets/${testAddress}/transactions?limit=5`);
      console.log('✅ Wallet transactions retrieved');
      console.log('   Transactions:', transactionsResponse.data.data?.length || 0);
    } catch (error) {
      console.log('⚠️  Wallet transactions failed (expected without API keys):', error.response?.status);
    }

    // Test 8: Invalid address validation
    console.log('\n8. Testing address validation...');
    try {
      await axios.get(`${API_BASE_URL}/api/wallets/invalid-address`);
      console.log('❌ Address validation failed - should have rejected invalid address');
    } catch (error) {
      if (error.response?.status === 400) {
        console.log('✅ Address validation working - rejected invalid address');
      } else {
        console.log('⚠️  Unexpected error:', error.response?.status);
      }
    }

    console.log('\n🎉 API testing completed!');
    console.log('\n📝 Notes:');
    console.log('   - Some endpoints may fail without proper database setup');
    console.log('   - Ethereum integration requires RPC endpoints and API keys');
    console.log('   - This is expected for initial testing');

  } catch (error) {
    console.error('❌ API test failed:', error.message);
    if (error.code === 'ECONNREFUSED') {
      console.log('\n💡 Make sure the backend server is running:');
      console.log('   cd crypto-bubble-map-be && npm run dev');
    }
  }
}

// Run the test
testAPI();
