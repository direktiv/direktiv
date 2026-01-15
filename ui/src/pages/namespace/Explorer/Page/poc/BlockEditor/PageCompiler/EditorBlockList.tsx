import {
  BlockListProps,
  BlockListWrapper,
} from "../../PageCompiler/BlockList/LiveBlockList";
import { Suspense, useMemo } from "react";

import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { Loading } from "../../PageCompiler/Block/utils/Loading";
import { useValidateDropzone } from "./useValidateDropzone";
import { useVariablesContext } from "../../PageCompiler/primitives/Variable/VariableContext";

export const EditorBlockList = ({
  horizontal,
  children,
  path,
}: BlockListProps) => {
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
