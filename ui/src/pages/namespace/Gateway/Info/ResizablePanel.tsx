import React, { ReactNode, useRef } from "react";
import {
  useLeftPanelWidth,
  usePanelSizeActions,
} from "../../../../util/store/panelSize";

import { useResizeDrag } from "./useResizeDrag";

type ResizablePanelProps = {
  leftPanel: ReactNode;
  rightPanel: ReactNode;
};

const ResizablePanel: React.FC<ResizablePanelProps> = ({
  leftPanel,
  rightPanel,
}) => {
  const leftPanelWidth = useLeftPanelWidth();
  const { setLeftPanelWidth } = usePanelSizeActions();
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
    <div ref={containerRef} className="w-full flex-col lg:flex lg:flex-row">
      {/* Left panel */}
      <div
        className="w-full max-lg:!w-full"
        style={{ width: `${leftPanelWidth}%` }}
      >
        {leftPanel}
      </div>

      {/* Resize handle - only visible on lg screens and above */}
      <div
        className="mx-2 hidden min-h-full w-1 shrink-0 cursor-col-resize hover:bg-gray-4 dark:hover:bg-gray-4 lg:block"
        onMouseDown={startResize}
      />

      {/* Right panel */}
      <div
        className="mt-4 max-lg:!w-full lg:mt-0"
        style={{ width: `${100 - leftPanelWidth - 0.25}%` }}
      >
        {rightPanel}
      </div>
    </div>
  );
};

export default ResizablePanel;
