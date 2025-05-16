import { AllBlocksType } from "../../schema/blocks";
import { BlockPath } from "./utils/blockPath";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Button } from "./Button";
import { Card } from "./Card";
import { Columns } from "./Columns";
import { Dialog } from "./Dialog";
import { Headline } from "./Headline";
import { Loop } from "./Loop";
import { ParsingError } from "./utils/ParsingError";
import { QueryProvider } from "./QueryProvider";
import { Table } from "./Table";
import { Text } from "./Text";
import { useTranslation } from "react-i18next";

type BlockProps = {
  block: AllBlocksType;
  blockPath: BlockPath;
};

export const Block = ({ block, blockPath }: BlockProps) => {
  const { t } = useTranslation();
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
      case "table":
        return <Table blockProps={block} />;
      default:
        return (
          <ParsingError
            title={t("direktivPage.error.blockNotImplemented", {
              type: block.type,
            })}
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
