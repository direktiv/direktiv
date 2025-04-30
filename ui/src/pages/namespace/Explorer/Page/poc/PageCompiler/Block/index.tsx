import { AllBlocksType } from "../../schema/blocks";
import { BlockPath } from "./utils/blockPath";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Button } from "./Button";
import { Card } from "./Card";
import { Dialog } from "./Dialog";
import { Headline } from "./Headline";
import { Loop } from "./Loop";
import { QueryProvider } from "./QueryProvider";
import { Text } from "./Text";
import { TwoColumns } from "./TwoColumns";
import { UserError } from "./utils/UserError";
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
      case "two-columns":
        return <TwoColumns blockProps={block} blockPath={blockPath} />;
      case "loop":
        return <Loop blockProps={block} blockPath={blockPath} />;
      case "query-provider":
        return <QueryProvider blockProps={block} blockPath={blockPath} />;
      case "dialog":
        return <Dialog blockProps={block} blockPath={blockPath} />;
      default:
        return (
          <UserError
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
