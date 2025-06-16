const axios = require('axios');

async function quickTest() {
  console.log('🧪 Quick API Test...\n');

  try {
    // Test the wallet network endpoint that was failing
    const testAddress = '0x1f9840a85d5af5bf1d1762f925bdaddc4201f984'; // UNI token contract
    console.log(`Testing wallet network for: ${testAddress}`);
    
    const response = await axios.get(`http://localhost:3001/api/wallets/network?address=${testAddress}&depth=2`);
    
    console.log('✅ Success!');
    console.log('Status:', response.status);
    console.log('Response:', {
      success: response.data.success,
      nodes: response.data.data?.nodes?.length || 0,
      links: response.data.data?.links?.length || 0,
      timestamp: response.data.timestamp
    });

  } catch (error) {
    console.log('❌ Error:', error.response?.status, error.response?.data?.error || error.message);
  }
}

quickTest();
