import Alert from "~/design/Alert";
import { AllBlocksType } from "../../schema/blocks";
import { BlockPath } from "./utils/blockPath";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Dialog } from "./Dialog";
import { Headline } from "./Headline";
import { QueryProvider } from "./QueryProvider";
import { Text } from "./Text";
import { TwoColumns } from "./TwoColumns";

type BlockProps = {
  block: AllBlocksType;
  blockPath: BlockPath;
};

export const Block = ({ block, blockPath }: BlockProps) => {
  const renderContent = () => {
    switch (block.type) {
      case "headline":
        return <Headline blockProps={block} />;
      case "text":
        return <Text blockProps={block} />;
      case "two-columns":
        return <TwoColumns blockProps={block} blockPath={blockPath} />;
      case "query-provider":
        return <QueryProvider blockProps={block} blockPath={blockPath} />;
      case "dialog":
        return <Dialog blockProps={block} blockPath={blockPath} />;
      default:
        return (
          <Alert variant="warning">not implemented yet: {block.type}</Alert>
        );
    }
  };

  return <BlockWrapper blockPath={blockPath}>{renderContent()}</BlockWrapper>;
};
