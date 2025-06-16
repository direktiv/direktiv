import { Block, BlockPathType } from ".";

import { BlockList } from "./utils/BlockList";
import Button from "~/design/Button";
import { FormType } from "../../schema/blocks/form";
import { usePageMutation } from "../procedures/mutation";
import { useTranslation } from "react-i18next";

type FormProps = {
  blockProps: FormType;
  blockPath: BlockPathType;
};

export const Form = ({ blockProps, blockPath }: FormProps) => {
  const { mutation } = blockProps;

  const { mutate } = usePageMutation(mutation);

  const { t } = useTranslation();
  return (
    <form
      id={mutation.id}
      onSubmit={(e) => {
        e.preventDefault();
        mutate();
      }}
    >
      <BlockList>
        {blockProps.blocks.map((block, index) => (
          <Block key={index} block={block} blockPath={[...blockPath, index]} />
        ))}
      </BlockList>
      <div className="mt-4 flex justify-end">
        <Button>{t("direktivPage.page.blocks.form.save")}</Button>
      </div>
    </form>
  );
};
