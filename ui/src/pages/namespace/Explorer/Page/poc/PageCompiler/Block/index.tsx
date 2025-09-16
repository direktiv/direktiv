import { BlockType } from "../../schema/blocks";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Button } from "./Button";
import { Card } from "./Card";
import { Columns } from "./Columns";
import { Dialog } from "./Dialog";
import { Form } from "./Form";
import { FormCheckbox } from "./formPrimitives/FormCheckbox";
import { FormDateInput } from "./formPrimitives/FormDateInput";
import { FormNumberInput } from "./formPrimitives/FormNumberInput";
import { FormSelect } from "./formPrimitives/FormSelect";
import { FormStringInput } from "./formPrimitives/FormStringInput";
import { FormTextarea } from "./formPrimitives/FormTextarea";
import { Headline } from "./Headline";
import { Image } from "./Image";
import { Loop } from "./Loop";
import { ParsingError } from "./utils/ParsingError";
import { QueryProvider } from "./QueryProvider";
import { Table } from "./Table";
import { Text } from "./Text";
import { useTranslation } from "react-i18next";

export type BlockPathType = number[];

type BlockProps = {
  block: BlockType;
  blockPath: BlockPathType;
  register?: (fields: string[]) => void;
};

export const Block = ({ block, blockPath, register }: BlockProps) => {
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
        return (
          <Form blockProps={block} blockPath={blockPath} register={register} />
        );
      case "form-string-input":
        return <FormStringInput blockProps={block} />;
      case "form-number-input":
        return <FormNumberInput blockProps={block} />;
      case "form-date-input":
        return <FormDateInput blockProps={block} />;
      case "form-textarea":
        return <FormTextarea blockProps={block} />;
      case "form-select":
        return <FormSelect blockProps={block} />;
      case "form-checkbox":
        return <FormCheckbox blockProps={block} />;
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
