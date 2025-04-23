import { PropsWithChildren, Suspense } from "react";

import { BlockPath } from "./blockPath";
import { Loading } from "./Loading";

/**
 * TODO:
 * [] a path to the wrapper
 */

type BlockWrapperProps = {
  blockPath: BlockPath;
} & PropsWithChildren;
export const BlockWrapper = ({ children, blockPath }: BlockWrapperProps) => (
  <div className="border p-3 border-dashed relative group">
    <div className="absolute  group-hover:block text-sm bg-slate-400 rounded-full px-2">
      {blockPath}
    </div>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </div>
);
