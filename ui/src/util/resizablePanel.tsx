import React, { ReactNode, useRef } from "react";
import { useLeftPanelWidth, useSetLeftPanelWidth } from "./store/panelSize";

import { useResizeDrag } from "../hooks/useResizeDrag";

type ResizablePanelProps = {
  leftPanel: ReactNode;
  rightPanel: ReactNode;
};

const ResizablePanel: React.FC<ResizablePanelProps> = ({
  leftPanel,
  rightPanel,
}) => {
  const leftPanelWidth = useLeftPanelWidth();
  const setLeftPanelWidth = useSetLeftPanelWidth();
  const containerRef = useRef<HTMLDivElement>(null);

  const minLeftWidth = 30;
  const maxLeftWidth = 70;

  const startResize = useResizeDrag({
    minLeftWidth,
    maxLeftWidth,
    onResize: setLeftPanelWidth,
    containerRef,
  });

  return (
    <div ref={containerRef} className="lg:flex lg:flex-row flex-col w-full">
      {/* Left panel */}
      <div
        className="max-lg:!w-full w-full"
        style={{ width: `${leftPanelWidth}%` }}
      >
        {leftPanel}
      </div>

      {/* Resize handle - only visible on lg screens and above */}
      <div
        className="w-1 min-h-full hover:bg-gray-4 dark:hover:bg-gray-4 cursor-col-resize shrink-0 mx-2 hidden lg:block"
        onMouseDown={startResize}
      />

      {/* Right panel */}
      <div
        className="lg:mt-0 max-lg:!w-full mt-4"
        style={{ width: `${100 - leftPanelWidth - 0.25}%` }}
      >
        {rightPanel}
      </div>
    </div>
  );
};

export default ResizablePanel;
