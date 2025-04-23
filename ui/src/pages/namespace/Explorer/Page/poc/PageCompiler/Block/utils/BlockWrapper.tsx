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
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (containerRef.current) {
        const allBlockWrapper = Array.from(
          document.querySelectorAll("[data-block-wrapper]")
        ).filter((element) => element.contains(e.target as Node));

        const deepestChildren = allBlockWrapper.at(-1);
        setIsHovered(containerRef.current === deepestChildren);
      }
    };

    document.addEventListener("mousemove", handleMouseMove);
    return () => document.removeEventListener("mousemove", handleMouseMove);
  }, []);

  return (
    <div
      ref={containerRef}
      className="relative border border-dashed p-3"
      data-block-wrapper
    >
      <Badge
        className="-m-6 absolute"
        variant="secondary"
        style={{
          display: isHovered ? "block" : "none",
        }}
      >
        {blockPath}
      </Badge>
      <Suspense fallback={<Loading />}>{children}</Suspense>
    </div>
  );
};
