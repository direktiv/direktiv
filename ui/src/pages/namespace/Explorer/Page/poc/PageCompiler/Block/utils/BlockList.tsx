import { ReactElement, Suspense, useMemo } from "react";

import { BlockPathType } from "..";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";
import { usePageStateContext } from "../../context/pageCompilerContext";
import { useValidateDropzone } from "./useValidateDropzone";
import { useVariablesContext } from "../../primitives/Variable/VariableContext";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

type WrapperProps = {
  horizontal?: boolean;
  children: ReactElement;
};

const BlockListWrapper = ({ children, horizontal }: WrapperProps) => (
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

const EditorBlockList = ({ horizontal, children, path }: BlockListProps) => {
  const validateDropzone = useValidateDropzone();
  const variables = useVariablesContext();

  const dropzonePayload = useMemo(
    () => ({ targetPath: [...path, 0], variables }),
    [path, variables]
  );

  return (
    <BlockListWrapper horizontal={horizontal}>
      <Suspense fallback={<Loading />}>
        {!children.length && (
          <div className="flex h-full min-h-[25px] flex-col justify-center">
            <Dropzone validate={validateDropzone} payload={dropzonePayload} />
          </div>
        )}
        {children}
      </Suspense>
    </BlockListWrapper>
  );
};

export const VisitorBlockList = ({ horizontal, children }: BlockListProps) => (
  <BlockListWrapper horizontal={horizontal}>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </BlockListWrapper>
);

export const BlockList = (props: BlockListProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockList {...props} />;
  }

  return <VisitorBlockList {...props} />;
};
