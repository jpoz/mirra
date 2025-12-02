import React from "react";
import { format } from "date-fns";
import { Recording } from "@/lib/api";
import { getStatusColor, getProviderStyles } from "@/lib/styles";

interface RecordingMetadataProps {
  recording: Recording;
}

/**
 * Displays key metadata: timestamp, provider, duration, status
 */
export function RecordingMetadata({ recording }: RecordingMetadataProps) {
  return (
    <div className="grid grid-cols-4 gap-4 mb-6">
      <div>
        <label className="text-sm font-medium text-muted-foreground">
          Timestamp
        </label>
        <p className="text-sm mt-1">
          {format(new Date(recording.timestamp), "MMM d, yyyy HH:mm:ss")}
        </p>
      </div>
      <div>
        <label className="text-sm font-medium text-muted-foreground">
          Provider
        </label>
        <p className="text-sm mt-1">
          <span
            className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getProviderStyles(recording.provider)}`}
          >
            {recording.provider}
          </span>
        </p>
      </div>
      <div>
        <label className="text-sm font-medium text-muted-foreground">
          Duration
        </label>
        <p className="text-sm mt-1">{recording.timing.duration_ms}ms</p>
      </div>
      <div>
        <label className="text-sm font-medium text-muted-foreground">
          Status
        </label>
        <p className="text-sm mt-1">
          <span
            className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getStatusColor(recording.response.status)}`}
          >
            {recording.response.status}
          </span>
        </p>
      </div>
    </div>
  );
}
