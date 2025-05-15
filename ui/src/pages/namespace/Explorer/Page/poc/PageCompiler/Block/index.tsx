import { AllBlocksType } from "../../schema/blocks";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Button } from "./Button";
import { Card } from "./Card";
import { Columns } from "./Columns";
import { Dialog } from "./Dialog";
import { Headline } from "./Headline";
import { Loop } from "./Loop";
import { ParsingError } from "./utils/ParsingError";
import { QueryProvider } from "./QueryProvider";
import { Text } from "./Text";

type BlockProps = {
  block: AllBlocksType;
  blockPath: BlockPath;
};

export type BlockPath = number[];

export const Block = ({ block, blockPath }: BlockProps) => {
  const renderContent = () => {
    switch (block.type) {
      case "headline":
        return <Headline blockProps={block} />;
      case "text":
        return <Text blockProps={block} />;
      case "card":
        return <Card blockProps={block} blockPath={blockPath} />;
      case "button":
        return <Button blockProps={block} />;
      case "columns":
        return <Columns blockProps={block} blockPath={blockPath} />;
      case "loop":
        return <Loop blockProps={block} blockPath={blockPath} />;
      case "query-provider":
        return <QueryProvider blockProps={block} blockPath={blockPath} />;
      case "dialog":
        return <Dialog blockProps={block} blockPath={blockPath} />;
      default:
        return (
          <ParsingError
            title={`The block type ${block.type} is not implemented yet`}
          />
        );
    }
  };

  return (
    <BlockWrapper blockPath={blockPath} block={block}>
      {renderContent()}
    </BlockWrapper>
  );
};
