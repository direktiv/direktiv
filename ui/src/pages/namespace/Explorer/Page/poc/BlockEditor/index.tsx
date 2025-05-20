import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Text } from "../BlockEditor/Text";
import { useBlock } from "../PageCompiler/context/pageCompilerContext";

export type BlockFormProps = { path: BlockPathType };

export type BlockEditFormProps = { block: AllBlocksType; path: BlockPathType };

export const BlockForm = ({ path }: { path: BlockPathType }) => {
  const block = useBlock(path);

  if (Array.isArray(block)) {
    throw Error("Can not load list into block editor");
  }

  switch (block.type) {
    case "text": {
      return <Text block={block} path={path} />;
    }
  }

  return (
    <div>
      Block form for {path} from {JSON.stringify(block)}
    </div>
  );
};
