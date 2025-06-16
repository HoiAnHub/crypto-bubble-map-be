#!/usr/bin/env node

/**
 * Test the exact Redis configuration that the application is using
 */

require('dotenv').config();

// Import the config exactly as the app does
const path = require('path');
const tsNode = require('ts-node');

// Register TypeScript
tsNode.register({
  project: path.join(__dirname, 'tsconfig.json'),
  transpileOnly: true,
});

// Now we can import TypeScript modules
const { config } = require('./src/config/config.ts');
const { createClient } = require('redis');

async function testAppRedisConfig() {
  console.log('🔧 Testing Application Redis Configuration');
  console.log('==========================================\n');

  console.log('Environment Variables:');
  console.log(`   REDIS_HOST: ${process.env.REDIS_HOST || 'undefined'}`);
  console.log(`   REDIS_PORT: ${process.env.REDIS_PORT || 'undefined'}`);
  console.log(`   REDIS_PASS: ${process.env.REDIS_PASS ? '***' : 'undefined'}`);
  console.log(`   REDIS_DB: ${process.env.REDIS_DB || 'undefined'}`);
  console.log(`   REDIS_URL: ${process.env.REDIS_URL || 'undefined'}\n`);

  console.log('Parsed Configuration:');
  console.log(`   config.redis.host: ${config.redis.host}`);
  console.log(`   config.redis.port: ${config.redis.port}`);
  console.log(`   config.redis.password: ${config.redis.password ? '***' : 'undefined'}`);
  console.log(`   config.redis.db: ${config.redis.db}`);
  console.log(`   config.redis.url: ${config.redis.url || 'undefined'}\n`);

  // Test the exact configuration the app would use
  let redisConfig;
  
  if (config.redis.url) {
    redisConfig = {
      url: config.redis.url,
    };
    console.log('Using URL-based configuration');
  } else {
    redisConfig = {
      socket: {
        host: config.redis.host,
        port: config.redis.port,
        connectTimeout: 5000,
        keepAlive: 5000,
        noDelay: true,
      },
      password: config.redis.password,
      database: config.redis.db,
    };
    console.log('Using individual parameter configuration');
  }

  console.log('\nTesting Redis connection with app configuration...');
  console.log('Configuration:', JSON.stringify(redisConfig, null, 2));

  let client = null;
  try {
    client = createClient(redisConfig);
    
    // Set up error handler
    client.on('error', (error) => {
      console.log(`❌ Redis Error: ${error.message}`);
    });

    client.on('connect', () => {
      console.log('✅ Redis connected event fired');
    });

    client.on('ready', () => {
      console.log('✅ Redis ready event fired');
    });

    console.log('Attempting to connect...');
    await client.connect();
    
    console.log('✅ Connection successful!');
    
    // Test basic operations
    await client.ping();
    console.log('✅ Ping successful');
    
    await client.set('test:app-config', 'working');
    const value = await client.get('test:app-config');
    console.log(`✅ Set/Get successful (value: ${value})`);
    
    await client.del('test:app-config');
    console.log('✅ Cleanup successful');
    
    await client.quit();
    console.log('✅ Disconnection successful');
    
    console.log('\n🎉 SUCCESS: Application Redis configuration is working!');
    
  } catch (error) {
    console.log(`\n❌ FAILED: ${error.message}`);
    console.log('\nPossible issues:');
    console.log('1. Redis server is not running');
    console.log('2. Wrong password configuration');
    console.log('3. Network connectivity issues');
    console.log('4. Redis server requires different authentication');
    
    if (client) {
      try {
        await client.quit();
      } catch (e) {
        // Ignore cleanup errors
      }
    }
  }
}

if (require.main === module) {
  testAppRedisConfig().catch(console.error);
}
