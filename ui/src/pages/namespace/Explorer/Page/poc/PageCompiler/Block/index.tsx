import { AllBlocksType } from "../../schema/blocks";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Button } from "./Button";
import { Card } from "./Card";
import { Columns } from "./Columns";
import { Dialog } from "./Dialog";
import { Form } from "./Form";
import { Headline } from "./Headline";
import { Image } from "./Image";
import { Loop } from "./Loop";
import { ParsingError } from "./utils/ParsingError";
import { QueryProvider } from "./QueryProvider";
import { Table } from "./Table";
import { Text } from "./Text";
import { useTranslation } from "react-i18next";

type BlockProps = {
  block: AllBlocksType;
  blockPath: BlockPathType;
};

export type BlockPathType = number[];

export const Block = ({ block, blockPath }: BlockProps) => {
  const { t } = useTranslation();
  const renderContent = () => {
    switch (block.type) {
      case "headline":
        return <Headline blockProps={block} />;
      case "text":
        return <Text blockProps={block} />;
      case "image":
        return <Image blockProps={block} />;
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
      case "form":
        return <Form blockProps={block} blockPath={blockPath} />;
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
