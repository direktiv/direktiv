import { useCallback, useEffect, useRef, useState } from "react";

/**
 * Hook that tracks an element's bounding rect and updates on window resize.
 *
 * Uses useRef for rect storage to avoid triggering re-renders when the ref
 * callback is called. This prevents infinite update loops that would occur
 * if setState was called in the ref callback.
 *
 * Uses useState (forceUpdate) to trigger re-renders only when the window
 * resizes, ensuring the component reflects the updated rect values.
 */
export const useElementRect = () => {
  const elementRef = useRef<HTMLDivElement | null>(null);
  const rectRef = useRef<DOMRect | undefined>(undefined);
  const [, forceUpdate] = useState({});

  const setRectCallback = useCallback((el: HTMLDivElement | null) => {
    if (el && !elementRef.current) {
      elementRef.current = el;
      rectRef.current = el.getBoundingClientRect();
    }
  }, []);

  useEffect(() => {
    const handleResize = () => {
      if (elementRef.current) {
        rectRef.current = elementRef.current.getBoundingClientRect();
        forceUpdate({});
      }
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return { ref: setRectCallback, rect: rectRef.current };
};
