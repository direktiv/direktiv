import {
  PageCompilerMode,
  usePageEditor,
} from "../../context/pageCompilerContext";
import { ReactElement, Suspense } from "react";

import { Loading } from "./Loading";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { getBlockTemplate } from "../../context/utils";
import { twMergeClsx } from "~/util/helpers";
import { useBlockDialog } from "../../../BlockEditor/BlockDialogProvider";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
};

type BlockListComponentProps = BlockListProps & { mode?: PageCompilerMode };

export const BlockList = ({
  horizontal,
  children,
}: BlockListComponentProps) => {
  const { setDialog } = useBlockDialog();
  const { mode } = usePageEditor();

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
              onSelect={(type) =>
                setDialog({
                  action: "create",
                  block: getBlockTemplate(type),
                  path: [0],
                })
              }
            />
          </div>
        )}
        {children}
      </Suspense>
    </div>
  );
};
