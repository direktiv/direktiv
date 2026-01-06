import { useEffect, useState } from "react";

export const useParentRect = (container: HTMLElement | null) => {
  const [rect, setRect] = useState<DOMRect | null>(null);

  useEffect(() => {
    const updateRect = () => {
      setRect(container?.parentElement?.getBoundingClientRect() || null);
    };
    updateRect();
    window.addEventListener("resize", updateRect);

    return () => window.removeEventListener("resize", updateRect);
  }, [container]);

  return rect;
};
