"use client";

import React from "react";
import { Routes, Route } from "react-router";
import Recordings from "./pages/Recordings";

function App() {
  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4">
          <div className="flex h-16 items-center justify-between">
            <div className="flex items-center">
              <a
                href="/"
                className="flex text-xl font-bold libertinus-math-regular"
              >
                <span className="ml-2 mt-3">𝕄𝕀ℝℝ𝔸</span>
              </a>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto flex-1 p-4">
        <Routes>
          <Route path="/" element={<Recordings />} />
        </Routes>
      </main>

      {/* Footer */}
      <footer className="border-t py-6">
        <div className="container mx-auto px-4 text-center text-sm text-muted-foreground h-8"></div>
      </footer>
    </div>
  );
}

export default App;
