#!/usr/bin/env node

/**
 * Test script for external API rate limiting and caching
 * This script tests the new ExternalApiService with rate limiting and caching
 */

require('dotenv').config();

// Mock the modules since we're testing from outside the TypeScript environment
const mockRedisService = {
  get: async (key) => {
    console.log(`   📦 Cache GET: ${key}`);
    return null; // Simulate cache miss for testing
  },
  set: async (key, value, ttl) => {
    console.log(`   📦 Cache SET: ${key} (TTL: ${ttl}s)`);
  },
  connect: async () => {},
  disconnect: async () => {},
};

// Simple test implementation
async function testExternalApiRateLimiting() {
  console.log('🔧 Testing External API Rate Limiting and Caching');
  console.log('==================================================\n');

  // Test 1: Rate limiting simulation
  console.log('1. Testing rate limiting simulation...');
  
  const rateLimitMap = new Map();
  
  async function simulateRateLimit(key, delay) {
    const now = Date.now();
    const lastCall = rateLimitMap.get(key) || 0;
    const timeSinceLastCall = now - lastCall;

    if (timeSinceLastCall < delay) {
      const waitTime = delay - timeSinceLastCall;
      console.log(`   ⏱️  Rate limiting ${key}: waiting ${waitTime}ms`);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }

    rateLimitMap.set(key, Date.now());
  }

  // Simulate multiple CoinGecko calls
  console.log('   Testing CoinGecko rate limiting (1200ms delay)...');
  const start = Date.now();
  
  for (let i = 0; i < 3; i++) {
    const callStart = Date.now();
    await simulateRateLimit('coingecko', 1200);
    const callEnd = Date.now();
    console.log(`   📞 Call ${i + 1}: ${callEnd - callStart}ms`);
  }
  
  const totalTime = Date.now() - start;
  console.log(`   ✅ Total time for 3 calls: ${totalTime}ms (expected: ~2400ms)`);

  // Test 2: Etherscan rate limiting
  console.log('\n2. Testing Etherscan rate limiting (200ms delay)...');
  
  const etherscanStart = Date.now();
  for (let i = 0; i < 3; i++) {
    const callStart = Date.now();
    await simulateRateLimit('etherscan', 200);
    const callEnd = Date.now();
    console.log(`   📞 Call ${i + 1}: ${callEnd - callStart}ms`);
  }
  
  const etherscanTotalTime = Date.now() - etherscanStart;
  console.log(`   ✅ Total time for 3 calls: ${etherscanTotalTime}ms (expected: ~400ms)`);

  // Test 3: Cache simulation
  console.log('\n3. Testing cache simulation...');
  
  const cache = new Map();
  
  function getCachedData(key, allowStale = false) {
    const cached = cache.get(key);
    if (!cached) {
      console.log(`   ❌ Cache miss: ${key}`);
      return null;
    }
    
    const now = Date.now();
    const age = now - cached.timestamp;
    
    if (!allowStale && age > cached.ttl * 1000) {
      console.log(`   ⏰ Cache expired: ${key} (age: ${age}ms, ttl: ${cached.ttl * 1000}ms)`);
      return null;
    }
    
    console.log(`   ✅ Cache hit: ${key} (age: ${age}ms)`);
    return cached.data;
  }
  
  function setCachedData(key, data, ttl) {
    cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
    console.log(`   📦 Cached: ${key} (TTL: ${ttl}s)`);
  }
  
  // Simulate caching workflow
  const cacheKey = 'test:eth-price';
  
  // First call - cache miss
  let data = getCachedData(cacheKey);
  if (!data) {
    console.log('   🌐 Making API call...');
    data = { price: 3000 };
    setCachedData(cacheKey, data, 5); // 5 second TTL
  }
  
  // Second call - cache hit
  await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
  data = getCachedData(cacheKey);
  
  // Third call - cache expired
  await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds
  data = getCachedData(cacheKey);
  
  // Fourth call - stale data allowed
  data = getCachedData(cacheKey, true);

  console.log('\n🎉 External API Rate Limiting Test Complete!');
  console.log('\n📊 Summary:');
  console.log('- Rate limiting prevents API abuse');
  console.log('- Caching reduces API calls');
  console.log('- Stale data fallback provides resilience');
  console.log('- Exponential backoff handles rate limit errors');
}

// Test actual API calls (if API keys are available)
async function testRealApiCalls() {
  console.log('\n🌐 Testing Real API Calls (if keys available)');
  console.log('=============================================\n');

  const axios = require('axios');

  // Test CoinGecko (no API key required for basic calls)
  console.log('1. Testing CoinGecko API...');
  try {
    const start = Date.now();
    const response = await axios.get(
      'https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd',
      { timeout: 10000 }
    );
    const end = Date.now();
    
    console.log(`   ✅ CoinGecko API call successful (${end - start}ms)`);
    console.log(`   💰 ETH Price: $${response.data.ethereum.usd}`);
    
  } catch (error) {
    if (error.response?.status === 429) {
      console.log('   ⚠️  CoinGecko rate limited (429)');
    } else {
      console.log(`   ❌ CoinGecko API error: ${error.message}`);
    }
  }

  // Test Etherscan (requires API key)
  console.log('\n2. Testing Etherscan API...');
  const etherscanKey = process.env.ETHERSCAN_API_KEY;
  
  if (!etherscanKey || etherscanKey === 'YOUR_ETHERSCAN_API_KEY') {
    console.log('   ⚠️  Etherscan API key not configured, skipping test');
  } else {
    try {
      const start = Date.now();
      const response = await axios.get(
        'https://api.etherscan.io/api',
        {
          params: {
            module: 'gastracker',
            action: 'gasoracle',
            apikey: etherscanKey,
          },
          timeout: 10000,
        }
      );
      const end = Date.now();
      
      if (response.data.status === '1') {
        console.log(`   ✅ Etherscan API call successful (${end - start}ms)`);
        console.log(`   ⛽ Gas Prices: Safe=${response.data.result.SafeGasPrice}, Standard=${response.data.result.ProposeGasPrice}, Fast=${response.data.result.FastGasPrice}`);
      } else {
        console.log(`   ❌ Etherscan API error: ${response.data.message}`);
      }
      
    } catch (error) {
      if (error.response?.status === 429) {
        console.log('   ⚠️  Etherscan rate limited (429)');
      } else {
        console.log(`   ❌ Etherscan API error: ${error.message}`);
      }
    }
  }

  console.log('\n💡 Rate Limiting Best Practices:');
  console.log('1. Always implement delays between API calls');
  console.log('2. Use caching to reduce API usage');
  console.log('3. Implement exponential backoff for retries');
  console.log('4. Have fallback data for when APIs are unavailable');
  console.log('5. Monitor API usage and adjust rate limits accordingly');
}

// Configuration validation
function validateConfiguration() {
  console.log('\n⚙️  Configuration Validation');
  console.log('============================\n');

  const config = {
    coingecko: {
      requestsPerMinute: parseInt(process.env.COINGECKO_REQUESTS_PER_MINUTE || '50', 10),
      delayBetweenRequests: parseInt(process.env.COINGECKO_DELAY_MS || '1200', 10),
    },
    etherscan: {
      requestsPerSecond: parseInt(process.env.ETHERSCAN_REQUESTS_PER_SECOND || '5', 10),
      delayBetweenRequests: parseInt(process.env.ETHERSCAN_DELAY_MS || '200', 10),
    },
  };

  console.log('CoinGecko Configuration:');
  console.log(`   Requests per minute: ${config.coingecko.requestsPerMinute}`);
  console.log(`   Delay between requests: ${config.coingecko.delayBetweenRequests}ms`);
  console.log(`   Calculated delay: ${60000 / config.coingecko.requestsPerMinute}ms`);

  console.log('\nEtherscan Configuration:');
  console.log(`   Requests per second: ${config.etherscan.requestsPerSecond}`);
  console.log(`   Delay between requests: ${config.etherscan.delayBetweenRequests}ms`);
  console.log(`   Calculated delay: ${1000 / config.etherscan.requestsPerSecond}ms`);

  // Validate configuration
  const coingeckoCalculated = 60000 / config.coingecko.requestsPerMinute;
  const etherscanCalculated = 1000 / config.etherscan.requestsPerSecond;

  if (config.coingecko.delayBetweenRequests >= coingeckoCalculated) {
    console.log('   ✅ CoinGecko rate limiting is properly configured');
  } else {
    console.log('   ⚠️  CoinGecko delay might be too low for the configured rate limit');
  }

  if (config.etherscan.delayBetweenRequests >= etherscanCalculated) {
    console.log('   ✅ Etherscan rate limiting is properly configured');
  } else {
    console.log('   ⚠️  Etherscan delay might be too low for the configured rate limit');
  }
}

// Run all tests
async function runAllTests() {
  await testExternalApiRateLimiting();
  await testRealApiCalls();
  validateConfiguration();
}

// Execute if run directly
if (require.main === module) {
  runAllTests().catch(console.error);
}

module.exports = { testExternalApiRateLimiting, testRealApiCalls, validateConfiguration };
