#!/usr/bin/env deno run --allow-all

// Simple test runner with progress indicators

import { $ } from "jsr:@david/dax";

const start = Date.now();

console.log("🧪 Running tests with live progress...");
console.log("📊 Mode: Short tests only");
console.log("⚡ Parallel: 8 workers");
console.log("📝 Output: Verbose\n");
console.log("=" .repeat(50));

try {
  // Run tests with colored output
  await $`cd backend && go test -v -short -timeout 30s -parallel 8 ./...`.printCommand();
  
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log("\n" + "=" .repeat(50));
  console.log(`\n✅ All tests passed in ${elapsed}s`);
  console.log(`\n💡 Tip: Tests show as cached when code hasn't changed`);
} catch (error) {
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log("\n" + "=" .repeat(50));
  console.log(`\n❌ Tests failed after ${elapsed}s`);
  Deno.exit(1);
}