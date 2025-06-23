import { AllBlocksType, inlineBlockTypes } from "../../../schema/blocks";
import {
  PageCompilerMode,
  usePageEditor,
} from "../../context/pageCompilerContext";
import { ReactElement, Suspense } from "react";

import { BlockPathType } from "..";
import { Loading } from "./Loading";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { getBlockTemplate } from "../../context/utils";
import { twMergeClsx } from "~/util/helpers";
import { useBlockDialog } from "../../../BlockEditor/BlockDialogProvider";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
  restrict?: AllBlocksType["type"][];
};

type BlockListComponentProps = BlockListProps & { mode?: PageCompilerMode };

export const BlockList = ({
  horizontal,
  children,
  path,
}: BlockListComponentProps) => {
  const { setDialog } = useBlockDialog();
  const { mode, addBlock } = usePageEditor();

  return (
    <div
      className={twMergeClsx(
        "gap-3",
        horizontal
          ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
          : "flex flex-col"
      )}
    >
      <Suspense fallback={<Loading />}>
        {mode === "edit" && !children.length && (
          <div className="self-center">
            <SelectBlockType
              big
              path={path}
              onSelect={(type) => {
                if (inlineBlockTypes.includes(type)) {
                  return addBlock([...path, 0], getBlockTemplate(type), true);
                }
                setDialog({
                  action: "create",
                  block: getBlockTemplate(type),
                  path: [...path, 0],
                });
              }}
            />
          </div>
        )}
        {children}
      </Suspense>
    </div>
  );
};
