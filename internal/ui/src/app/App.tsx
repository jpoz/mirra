"use client";

import React from "react";
import { Routes, Route } from "react-router";
import Recordings from "./pages/Recordings";
import Recording from "./pages/Recording";
import { ModeToggle } from "./components/mode-toggle";

function App() {
  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* Header */}
      <header className="border-b">
        <div className=" mx-auto px-4">
          <div className="flex h-16 items-center justify-between">
            <div className="flex items-center">
              <a
                href="/"
                className="flex text-xl font-bold libertinus-math-regular"
              >
                <img src="/logo.png" alt="mirra" className="h-8 mr-2" />
              </a>
            </div>
            <div className="flex items-center gap-4">
              <ModeToggle />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main>
        <Routes>
          <Route path="/" element={<Recordings />} />
          <Route path="/recordings" element={<Recordings />} />
          <Route path="/recordings/:id" element={<Recording />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
