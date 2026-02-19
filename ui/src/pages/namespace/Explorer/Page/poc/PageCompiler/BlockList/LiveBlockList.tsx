import { ReactElement, Suspense } from "react";

import { BlockPathType } from "../Block";
import { Loading } from "../Block/utils/Loading";
import { twMergeClsx } from "~/util/helpers";

export type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

type WrapperProps = {
  horizontal?: boolean;
  children: ReactElement;
};

export const BlockListWrapper = ({ children, horizontal }: WrapperProps) => (
  <div
    className={twMergeClsx(
      "w-full",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))] gap-6"
        : "flex flex-col"
    )}
  >
    {children}
  </div>
);

export const LiveBlockList = ({ horizontal, children }: BlockListProps) => (
  <BlockListWrapper horizontal={horizontal}>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </BlockListWrapper>
);
