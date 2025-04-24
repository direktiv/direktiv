import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { UserError } from "./UserError";

type BlockWrapperProps = {
  blockPath: BlockPath;
  block: AllBlocksType;
} & PropsWithChildren;

export const BlockWrapper = ({
  children,
  block: { type },
  blockPath,
}: BlockWrapperProps) => {
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
        <b>{type}</b> {blockPath}
      </Badge>
      <Suspense fallback={<Loading />}>
        <ErrorBoundary
          fallbackRender={({ error }) => (
            <UserError title="There was an error fetching data from the API">
              {error.message}
            </UserError>
          )}
        >
          {children}
        </ErrorBoundary>
      </Suspense>
    </div>
  );
};
