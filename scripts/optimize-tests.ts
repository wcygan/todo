#!/usr/bin/env deno run --allow-all

import { walk } from "jsr:@std/fs/walk";

// Add t.Parallel() to all test functions
async function addParallelToTests(filePath: string) {
  const content = await Deno.readTextFile(filePath);
  const lines = content.split('\n');
  const updatedLines: string[] = [];
  
  for (let i = 0; i < lines.length; i++) {
    updatedLines.push(lines[i]);
    
    // Match test function definitions (with flexible whitespace)
    if (lines[i].match(/^func Test\w+\(t \*testing\.T\)\s*\{$/)) {
      // Check if next line isn't already t.Parallel()
      if (i + 1 < lines.length && !lines[i + 1].includes('t.Parallel()')) {
        updatedLines.push('\tt.Parallel()');
      }
    }
  }
  
  const updatedContent = updatedLines.join('\n');
  if (content !== updatedContent) {
    await Deno.writeTextFile(filePath, updatedContent);
    console.log(`✅ Added t.Parallel() to ${filePath}`);
    return true;
  }
  return false;
}

// Main execution
let filesUpdated = 0;
const testFiles = walk("backend/", {
  exts: ["_test.go"],
  includeDirs: false,
});

for await (const entry of testFiles) {
  if (await addParallelToTests(entry.path)) {
    filesUpdated++;
  }
}

console.log(`\n📊 Summary: Updated ${filesUpdated} test files with t.Parallel()`);