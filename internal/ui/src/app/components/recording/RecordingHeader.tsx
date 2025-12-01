import React from "react";
import { ArrowLeft } from "lucide-react";
import { useNavigate } from "react-router";

interface RecordingHeaderProps {
  recordingId: string;
}

/**
 * Header bar with back button and recording ID
 */
export function RecordingHeader({ recordingId }: RecordingHeaderProps) {
  const navigate = useNavigate();

  return (
    <div className="flex items-center justify-between p-6 border-b bg-card">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate("/recordings")}
          className="p-2 hover:bg-muted rounded-md transition-colors"
          aria-label="Back to recordings list"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div>
          <h2 className="text-xl font-bold">Recording Details</h2>
          <p className="text-sm text-muted-foreground font-mono mt-1">
            {recordingId}
          </p>
        </div>
      </div>
    </div>
  );
}
