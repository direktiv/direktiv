import { BlockListProps, LiveBlockList } from "./LiveBlockList";

import { EditorBlockList } from "../../BlockEditor/PageCompiler/EditorBlockList";
import { usePageStateContext } from "../context/pageCompilerContext";

export const BlockList = (props: BlockListProps) => {
  const { mode } = usePageStateContext();

  if (mode === "live") {
    return <LiveBlockList {...props} />;
  }

  return <EditorBlockList {...props} />;
};
