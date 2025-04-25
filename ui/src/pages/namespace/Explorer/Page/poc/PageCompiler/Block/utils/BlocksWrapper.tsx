import { PropsWithChildren, Suspense } from "react";

import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";

type BlocksWrapperProps = PropsWithChildren<{
  horizontal?: boolean;
}>;

export const BlocksWrapper = ({ horizontal, children }: BlocksWrapperProps) => (
  <div
    className={twMergeClsx(
      "gap-3",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
        : "flex flex-col"
    )}
  >
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </div>
);
