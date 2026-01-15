import { BlockWrapperProps, LiveBlockWrapper } from "./LiveBlockWrapper";

import { BlockWrapper as EditorBlockWrapper } from "../../BlockEditor/PageCompiler/EditorBlockWrapper";
import { usePageStateContext } from "../context/pageCompilerContext";

export const BlockWrapper = (props: BlockWrapperProps) => {
  const { mode } = usePageStateContext();

  if (mode === "live") {
    return <LiveBlockWrapper {...props} />;
  }

  return <EditorBlockWrapper {...props} />;
};
