import { useCallback, useEffect, useRef } from "react";

type UseResizeDragProps = {
  minLeftWidth: number;
  maxLeftWidth: number;
  onResize: (width: number) => void;
  containerRef: React.RefObject<HTMLDivElement>;
};

export const useResizeDrag = ({
  minLeftWidth,
  maxLeftWidth,
  onResize,
  containerRef,
}: UseResizeDragProps) => {
  const isDragging = useRef(false);

  const startResize = useCallback(() => {
    isDragging.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  }, []);

  useEffect(() => {
    const resize = (e: MouseEvent) => {
      if (isDragging.current && containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        const relativeX = e.clientX - rect.left;
        const percentage = (relativeX / rect.width) * 100;
        const newLeftWidth = Math.max(
          minLeftWidth,
          Math.min(maxLeftWidth, percentage)
        );

        onResize(newLeftWidth);
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
  }, [minLeftWidth, maxLeftWidth, onResize, containerRef]);

  return startResize;
};
