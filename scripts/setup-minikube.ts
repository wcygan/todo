#!/usr/bin/env -S deno run --allow-run --allow-env

/**
 * Setup minikube for Kubernetes development
 * Creates a safe development context for Tilt
 */

import { $ } from "jsr:@david/dax@0.42.0";

async function setupMinikube() {
  console.log("🚀 Setting up minikube for Kubernetes development...");

  try {
    // Check if minikube is installed
    console.log("📋 Checking minikube installation...");
    await $`minikube version`;
    
    // Start minikube if not running
    console.log("⚡ Starting minikube...");
    try {
      await $`minikube start --driver=docker`;
    } catch (error) {
      // If already running, that's fine
      console.log("ℹ️  Minikube may already be running, continuing...");
    }

    // Set kubectl context to minikube
    console.log("🔧 Setting kubectl context to minikube...");
    await $`kubectl config use-context minikube`;

    // Verify the context
    console.log("✅ Verifying kubectl context...");
    const context = await $`kubectl config current-context`.text();
    console.log(`Current context: ${context.trim()}`);

    // Enable required addons
    console.log("🔌 Enabling required minikube addons...");
    await $`minikube addons enable ingress`;
    await $`minikube addons enable dashboard`;
    await $`minikube addons enable metrics-server`;

    // Get minikube IP for reference
    console.log("🌐 Getting minikube IP...");
    const ip = await $`minikube ip`.text();
    console.log(`Minikube IP: ${ip.trim()}`);

    console.log("✅ Minikube setup complete!");
    console.log("🎯 You can now run 'tilt up' safely with the minikube context");
    console.log("📊 Access minikube dashboard with: minikube dashboard");
    
  } catch (error) {
    console.error("❌ Error setting up minikube:", error);
    Deno.exit(1);
  }
}

if (import.meta.main) {
  await setupMinikube();
}