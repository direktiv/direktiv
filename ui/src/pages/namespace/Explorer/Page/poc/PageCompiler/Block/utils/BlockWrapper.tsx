import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";

import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import { Loading } from "./Loading";

type BlockWrapperProps = {
  blockPath: BlockPath;
} & PropsWithChildren;

export const BlockWrapper = ({ children, blockPath }: BlockWrapperProps) => {
  const [isHovered, setIsHovered] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (wrapperRef.current) {
        const containingWrappers = Array.from(
          document.querySelectorAll("[data-block-wrapper]")
        ).filter((wrapper) => wrapper.contains(e.target as Node));

        // This wrapper is the deepest if it's the last one in the containing wrappers array
        const isDeepest =
          containingWrappers[containingWrappers.length - 1] ===
          wrapperRef.current;

        setIsHovered(!!isDeepest);
      }
    };

    document.addEventListener("mousemove", handleMouseMove);
    return () => document.removeEventListener("mousemove", handleMouseMove);
  }, []);

  // Only show badge if this wrapper is hovered and none of its children are
  const showBadge = isHovered;

  return (
    <div
      ref={wrapperRef}
      className="border p-3 border-dashed relative"
      data-block-wrapper
    >
      <Badge
        className="-m-6 absolute"
        variant="secondary"
        style={{
          display: showBadge ? "block" : "none",
        }}
      >
        {blockPath}
      </Badge>
      <Suspense fallback={<Loading />}>{children}</Suspense>
    </div>
  );
};
