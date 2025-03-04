import React, { ReactNode, useEffect, useRef, useState } from "react";

import { twMergeClsx } from "~/util/helpers";

type ResizablePanelProps = {
  leftPanel: ReactNode;
  rightPanel: ReactNode;
  initialLeftWidth?: number; // percentage (0-100)
  minLeftWidth?: number; // percentage (0-100)
  maxLeftWidth?: number; // percentage (0-100)
  handleClassName?: string;
  className?: string;
};

const ResizablePanel: React.FC<ResizablePanelProps> = ({
  leftPanel,
  rightPanel,
  initialLeftWidth = 85,
  minLeftWidth = 30,
  maxLeftWidth = 70,
  handleClassName,
  className,
}) => {
  const [leftWidth, setLeftWidth] = useState(initialLeftWidth);
  const isDragging = useRef(false);
  const [isLargeScreen, setIsLargeScreen] = useState(false);

  useEffect(() => {
    const resize = (e: MouseEvent) => {
      if (isDragging.current) {
        // Calculate percentage based on window width
        const percentage = (e.clientX / window.innerWidth) * 100;
        setLeftWidth(
          Math.max(minLeftWidth, Math.min(maxLeftWidth, percentage))
        );
      }
    };

    const stopResize = () => {
      isDragging.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    };

    document.addEventListener("mousemove", resize);
    document.addEventListener("mouseup", stopResize);

    return () => {
      document.removeEventListener("mousemove", resize);
      document.removeEventListener("mouseup", stopResize);
    };
  }, [minLeftWidth, maxLeftWidth]);

  useEffect(() => {
    const checkScreenSize = () => {
      // Purpose is to make component responsive and this adds breakpoint for lg screens
      setIsLargeScreen(window.innerWidth >= 1024);
    };

    checkScreenSize(); // Check initially
    window.addEventListener("resize", checkScreenSize);

    return () => window.removeEventListener("resize", checkScreenSize);
  }, []);

  const startResize = () => {
    isDragging.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  };

  return (
    <div
      className={twMergeClsx("lg:flex lg:flex-row flex-col w-full", className)}
    >
      {/* Left panel */}
      <div
        className="lg:w-auto w-full"
        style={isLargeScreen ? { width: `${leftWidth}%` } : {}}
      >
        {leftPanel}
      </div>

      {/* Resize handle - only visible on lg screens and above */}
      <div
        className={twMergeClsx(
          "w-1 min-h-full hover:bg-gray-100 cursor-col-resize shrink-0 mx-2 hidden lg:block",
          handleClassName
        )}
        onMouseDown={startResize}
      />

      {/* Right panel */}
      <div
        className="lg:w-auto w-full lg:mt-0 mt-4"
        style={isLargeScreen ? { width: `${100 - leftWidth - 0.25}%` } : {}}
      >
        {rightPanel}
      </div>
    </div>
  );
};

export default ResizablePanel;
