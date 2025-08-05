#!/usr/bin/env deno run --allow-all

// Fast test runner that optimizes for speed

import { $ } from "jsr:@david/dax";

const start = Date.now();

console.log("🚀 Running fast test suite...");
console.log("📊 Mode: Short tests only (skipping integration tests)");
console.log("⚡ Parallel: 8 workers");
console.log("⏱️  Timeout: 30 seconds\n");

// Run tests with optimal settings
try {
  // Change to backend directory and run tests with verbose output
  await $`cd backend && go test -v -short -timeout 30s -parallel 8 ./...`.printCommand();
  
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log(`\n✅ All tests passed in ${elapsed}s`);
} catch (error) {
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log(`\n❌ Tests failed after ${elapsed}s`);
  Deno.exit(1);
}