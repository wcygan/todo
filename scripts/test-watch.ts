#!/usr/bin/env deno run --allow-all

// Test watcher with real-time progress display

import { $ } from "jsr:@david/dax";

console.log("👀 Test Watcher - Shows test progress in real-time");
console.log("=" .repeat(50));

// Function to format test output
function formatTestLine(line: string): string {
  // Running a test
  if (line.includes("=== RUN")) {
    return `🏃 ${line}`;
  }
  // Test passed
  if (line.includes("--- PASS:")) {
    const match = line.match(/--- PASS: (\S+) \(([^)]+)\)/);
    if (match) {
      return `✅ ${match[1]} (${match[2]})`;
    }
  }
  // Test failed
  if (line.includes("--- FAIL:")) {
    const match = line.match(/--- FAIL: (\S+) \(([^)]+)\)/);
    if (match) {
      return `❌ ${match[1]} (${match[2]})`;
    }
  }
  // Test skipped
  if (line.includes("--- SKIP:")) {
    return `⏭️  ${line}`;
  }
  // Package result
  if (line.startsWith("ok") || line.startsWith("PASS")) {
    return `📦 ${line}`;
  }
  if (line.startsWith("FAIL")) {
    return `💥 ${line}`;
  }
  // Cache hit
  if (line.includes("(cached)")) {
    return `💾 ${line}`;
  }
  
  return line;
}

const start = Date.now();

try {
  console.log("\n🚀 Starting test run...\n");
  
  // Run tests with streaming output
  const command = $`cd backend && go test -v -short -timeout 30s -parallel 8 ./...`
    .stdout("piped")
    .stderr("piped");
  
  // Spawn the process
  const child = command.spawn();
  
  // Process stdout
  const decoder = new TextDecoder();
  const reader = child.stdout.getReader();
  
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    
    const text = decoder.decode(value);
    const lines = text.split('\n').filter(line => line.trim());
    
    for (const line of lines) {
      console.log(formatTestLine(line));
    }
  }
  
  // Wait for process to complete
  const status = await child.status;
  
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log("\n" + "=" .repeat(50));
  
  if (status.code === 0) {
    console.log(`✅ All tests passed in ${elapsed}s`);
  } else {
    console.log(`❌ Tests failed after ${elapsed}s`);
    Deno.exit(1);
  }
} catch (error) {
  const elapsed = ((Date.now() - start) / 1000).toFixed(2);
  console.log("\n" + "=" .repeat(50));
  console.log(`❌ Tests failed after ${elapsed}s`);
  
  if (error instanceof Error) {
    console.log(`\n💭 Error: ${error.message}`);
  }
  
  Deno.exit(1);
}