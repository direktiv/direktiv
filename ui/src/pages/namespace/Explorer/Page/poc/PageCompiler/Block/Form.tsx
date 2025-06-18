import { Block, BlockPathType } from ".";

import Alert from "~/design/Alert";
import { BlockList } from "./utils/BlockList";
import Button from "~/design/Button";
import { FormType } from "../../schema/blocks/form";
import { Send } from "lucide-react";
import { usePageMutation } from "../procedures/mutation";
import { useTranslation } from "react-i18next";

type FormProps = {
  blockProps: FormType;
  blockPath: BlockPathType;
};

export const Form = ({ blockProps, blockPath }: FormProps) => {
  const { mutation } = blockProps;

  const { mutate, isPending, error } = usePageMutation(mutation);

  const { t } = useTranslation();
  return (
    <form
      id={mutation.id}
      onSubmit={(e) => {
        e.preventDefault();
        mutate();
      }}
    >
      {error && (
        <Alert variant="error" className="mb-4">
          {error.message}
        </Alert>
      )}
      <BlockList>
        {blockProps.blocks.map((block, index) => (
          <Block key={index} block={block} blockPath={[...blockPath, index]} />
        ))}
      </BlockList>
      <div className="mt-4 flex justify-end">
        <Button disabled={isPending} loading={isPending}>
          {!isPending && <Send />}
          {t("direktivPage.page.blocks.form.save")}
        </Button>
      </div>
    </form>
  );
};
