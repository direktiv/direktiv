import { ReactElement, Suspense } from "react";

import { BlockPathType } from "..";
import { Loading } from "./Loading";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { twMergeClsx } from "~/util/helpers";
import { useCreateBlock } from "../../context/utils/useCreateBlock";
import { usePageStateContext } from "../../context/pageCompilerContext";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

type WrapperProps = {
  horizontal?: boolean;
  children: ReactElement;
};

type BlockListComponentProps = BlockListProps;

const BlockListWrapper = ({ children, horizontal }: WrapperProps) => (
  <div
    className={twMergeClsx(
      "w-full gap-3",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
        : "flex flex-col"
    )}
  >
    {children}
  </div>
);

const EditorBlockList = ({
  horizontal,
  children,
  path,
}: BlockListComponentProps) => {
  const createBlock = useCreateBlock();

  return (
    <BlockListWrapper horizontal={horizontal}>
      <Suspense fallback={<Loading />}>
        {!children.length && (
          <div
            className="self-center"
            onClick={(event) => event.stopPropagation()}
          >
            placeholder for droppable
          </div>
        )}
        {children}
      </Suspense>
    </BlockListWrapper>
  );
};

const VisitorBlockList = ({
  horizontal,
  children,
}: BlockListComponentProps) => (
  <BlockListWrapper horizontal={horizontal}>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </BlockListWrapper>
);

export const BlockList = (props: BlockListComponentProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockList {...props} />;
  }

  return <VisitorBlockList {...props} />;
};
