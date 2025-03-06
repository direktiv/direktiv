import React, { ReactNode, useRef } from "react";
import { usePanelSize, useSetPanelSize } from "./store/panelSize";

import { useResizeDrag } from "../hooks/useResizeDrag";

type ResizablePanelProps = {
  leftPanel: ReactNode;
  rightPanel: ReactNode;
  minLeftWidth?: number; // percentage (0-100)
  maxLeftWidth?: number; // percentage (0-100)
};

const ResizablePanel: React.FC<ResizablePanelProps> = ({
  leftPanel,
  rightPanel,
  minLeftWidth = 30,
  maxLeftWidth = 70,
}) => {
  const panelWidth = usePanelSize();
  const setPanelWidth = useSetPanelSize();
  const containerRef = useRef<HTMLDivElement>(null);

  const { startResize } = useResizeDrag({
    minWidth: minLeftWidth,
    maxWidth: maxLeftWidth,
    onResize: setPanelWidth,
    containerRef,
  });

  return (
    <div ref={containerRef} className="lg:flex lg:flex-row flex-col w-full">
      {/* Left panel */}
      <div
        className="max-lg:!w-full w-full"
        style={{ width: `${panelWidth}%` }}
      >
        {leftPanel}
      </div>

      {/* Resize handle - only visible on lg screens and above */}
      <div
        className="w-1 min-h-full hover:bg-gray-100 cursor-col-resize shrink-0 mx-2 hidden lg:block"
        onMouseDown={startResize}
      />

      {/* Right panel */}
      <div
        className="lg:mt-0 max-lg:!w-full mt-4"
        style={{ width: `${100 - panelWidth - 0.25}%` }}
      >
        {rightPanel}
      </div>
    </div>
  );
};

export default ResizablePanel;
