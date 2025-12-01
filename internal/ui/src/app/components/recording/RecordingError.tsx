import React from "react";

interface RecordingErrorProps {
  error: string;
}

/**
 * Displays error message in a styled alert box
 */
export function RecordingError({ error }: RecordingErrorProps) {
  return (
    <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md mb-6">
      <label className="text-sm font-medium text-red-800 dark:text-red-300">
        Error
      </label>
      <p className="text-sm text-red-700 dark:text-red-400 mt-1 font-mono">
        {error}
      </p>
    </div>
  );
}
