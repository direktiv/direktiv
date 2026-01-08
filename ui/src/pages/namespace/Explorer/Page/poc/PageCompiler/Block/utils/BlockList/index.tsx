import { ReactElement, Suspense } from "react";

import { BlockPathType } from "../..";
import { EditorBlockList } from "./EditorBlockList";
import { Loading } from "../Loading";
import { twMergeClsx } from "~/util/helpers";
import { usePageStateContext } from "../../../context/pageCompilerContext";

declare const __IS_PAGESAPP__: boolean;

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

export const VisitorBlockList = ({ horizontal, children }: BlockListProps) => (
  <BlockListWrapper horizontal={horizontal}>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </BlockListWrapper>
);

export const BlockList = (props: BlockListProps) => {
  const { mode } = usePageStateContext();

  if (__IS_PAGESAPP__ || mode === "live") {
    return <VisitorBlockList {...props} />;
  }

  return <EditorBlockList {...props} />;
};
