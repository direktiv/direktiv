import { PropsWithChildren, Suspense, useState } from "react";

import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import { Loading } from "./Loading";

type BlockWrapperProps = {
  blockPath: BlockPath;
} & PropsWithChildren;
export const BlockWrapper = ({ children, blockPath }: BlockWrapperProps) => {
  const [isHovered, setIsHovered] = useState(false);
  return (
    <div
      className="border p-3 border-dashed relative"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
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
