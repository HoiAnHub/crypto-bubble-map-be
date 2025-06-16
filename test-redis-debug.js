#!/usr/bin/env node

/**
 * Redis Connection Debug Script
 * Tests different Redis connection configurations to identify the issue
 */

const { createClient } = require('redis');

async function testRedisConnection() {
  console.log('🔧 Testing Redis Connection');
  console.log('============================\n');

  // Test configurations
  const configs = [
    {
      name: 'No Password',
      config: {
        socket: {
          host: 'localhost',
          port: 6379,
        }
      }
    },
    {
      name: 'With Password "dev"',
      config: {
        socket: {
          host: 'localhost',
          port: 6379,
        },
        password: 'dev'
      }
    },
    {
      name: 'URL without password',
      config: {
        url: 'redis://localhost:6379'
      }
    },
    {
      name: 'URL with password',
      config: {
        url: 'redis://:dev@localhost:6379'
      }
    }
  ];

  for (const { name, config } of configs) {
    console.log(`Testing: ${name}`);
    
    let client = null;
    try {
      client = createClient(config);
      
      // Set up error handler
      client.on('error', (error) => {
        console.log(`   ❌ Error: ${error.message}`);
      });

      // Connect with timeout
      const connectPromise = client.connect();
      const timeoutPromise = new Promise((_, reject) => 
        setTimeout(() => reject(new Error('Connection timeout')), 5000)
      );

      await Promise.race([connectPromise, timeoutPromise]);
      
      // Test a simple operation
      await client.ping();
      console.log(`   ✅ Success: Connected and ping successful`);
      
      // Test set/get operation
      await client.set('test:connection', 'working');
      const value = await client.get('test:connection');
      console.log(`   ✅ Success: Set/Get operation working (value: ${value})`);
      
      // Clean up test key
      await client.del('test:connection');
      
      await client.quit();
      console.log(`   ✅ Success: Disconnected cleanly\n`);
      
      // If we get here, this config works
      console.log(`🎉 WORKING CONFIGURATION FOUND: ${name}`);
      console.log('Configuration:', JSON.stringify(config, null, 2));
      return config;
      
    } catch (error) {
      console.log(`   ❌ Failed: ${error.message}\n`);
      
      if (client) {
        try {
          await client.quit();
        } catch (e) {
          // Ignore cleanup errors
        }
      }
    }
  }
  
  console.log('❌ No working configuration found!');
  return null;
}

// Test Redis info if we can connect
async function getRedisInfo() {
  console.log('\n🔍 Getting Redis Server Information');
  console.log('===================================\n');

  try {
    // Try the most basic connection first
    const client = createClient({
      socket: {
        host: 'localhost',
        port: 6379,
      }
    });

    await client.connect();
    
    // Get server info
    const info = await client.info();
    console.log('Redis Server Info:');
    
    // Parse relevant info
    const lines = info.split('\r\n');
    const relevantInfo = lines.filter(line => 
      line.includes('redis_version') ||
      line.includes('redis_mode') ||
      line.includes('role:') ||
      line.includes('connected_clients') ||
      line.includes('used_memory_human') ||
      line.includes('requirepass')
    );
    
    relevantInfo.forEach(line => {
      if (line.trim()) {
        console.log(`   ${line}`);
      }
    });

    await client.quit();
    
  } catch (error) {
    console.log(`❌ Could not get Redis info: ${error.message}`);
  }
}

// Main execution
async function main() {
  const workingConfig = await testRedisConnection();
  await getRedisInfo();
  
  if (workingConfig) {
    console.log('\n✅ SOLUTION FOUND!');
    console.log('==================\n');
    console.log('Update your .env file with the working configuration:');
    
    if (workingConfig.password) {
      console.log('REDIS_PASS=dev');
    } else {
      console.log('# Remove or comment out REDIS_PASS');
      console.log('# REDIS_PASS=dev');
    }
    
    if (workingConfig.url) {
      console.log(`REDIS_URL=${workingConfig.url}`);
    }
  }
  
  console.log('\n💡 Next Steps:');
  console.log('1. Update your .env file based on the working configuration above');
  console.log('2. Restart your application');
  console.log('3. Check for "✅ Redis connected successfully" in the logs');
}

if (require.main === module) {
  main().catch(console.error);
}
