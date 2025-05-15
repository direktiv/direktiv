import { BlockPath } from "../PageCompiler/Block";
import { useBlock } from "../PageCompiler/context/pageCompilerContext";

export const BlockForm = ({ path }: { path: BlockPath }) => {
  const block = useBlock(path);
  return (
    <div>
      Block form for {path} from {JSON.stringify(block)}
    </div>
  );
};
