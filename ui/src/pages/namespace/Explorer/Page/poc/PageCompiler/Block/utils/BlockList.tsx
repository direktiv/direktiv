import {
  PageCompilerMode,
  usePageEditor,
} from "../../context/pageCompilerContext";
import { ReactElement, Suspense } from "react";

import { BlockPathType } from "..";
import { Loading } from "./Loading";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { twMergeClsx } from "~/util/helpers";
import { useCreateBlock } from "../../context/utils/useCreateBlock";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

type BlockListComponentProps = BlockListProps & { mode?: PageCompilerMode };

export const BlockList = ({
  horizontal,
  children,
  path,
}: BlockListComponentProps) => {
  const { mode } = usePageEditor();
  const createBlock = useCreateBlock();

  return (
    <div
      className={twMergeClsx(
        "gap-3 px-2",
        horizontal
          ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
          : "flex flex-col"
      )}
    >
      <Suspense fallback={<Loading />}>
        {mode === "edit" && !children.length && (
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
