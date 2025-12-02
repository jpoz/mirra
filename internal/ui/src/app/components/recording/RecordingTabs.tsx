import React from "react";

interface RecordingTabsProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  tabs: Array<{ id: string; label: string }>;
}

/**
 * Tab navigation component
 */
export function RecordingTabs({
  activeTab,
  onTabChange,
  tabs,
}: RecordingTabsProps) {
  return (
    <div className="flex gap-6 border-b -mb-px">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          className={`
            pb-3 text-sm font-medium transition-colors border-b-2
            ${
              activeTab === tab.id
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
            }
          `}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
