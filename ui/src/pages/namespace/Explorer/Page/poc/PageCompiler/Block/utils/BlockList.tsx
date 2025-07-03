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

type BlockListComponentProps = BlockListProps;

const EditorBlockList = ({
  horizontal,
  children,
  path,
}: BlockListComponentProps) => {
  const createBlock = useCreateBlock();

  return (
    <div
      className={twMergeClsx(
        "w-full gap-3",
        horizontal
          ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
          : "flex flex-col"
      )}
    >
      <Suspense fallback={<Loading />}>
        {!children.length && (
          <div
            className="self-center"
            onClick={(event) => event.stopPropagation()}
          >
            <SelectBlockType
              big
              path={path}
              onSelect={(type) => createBlock(type, [...path, 0])}
            />
          </div>
        )}
        {children}
      </Suspense>
    </div>
  );
};

const VisitorBlockList = ({
  horizontal,
  children,
}: BlockListComponentProps) => (
  <div
    className={twMergeClsx(
      "w-full gap-3",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
        : "flex flex-col"
    )}
  >
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </div>
);

export const BlockList = (props: BlockListComponentProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockList {...props} />;
  }

  return <VisitorBlockList {...props} />;
};
