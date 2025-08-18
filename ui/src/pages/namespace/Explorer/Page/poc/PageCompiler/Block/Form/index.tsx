import { Block, BlockPathType } from "..";

import Alert from "~/design/Alert";
import { BlockList } from "../utils/BlockList";
import { Button } from "../Button";
import { FormType } from "../../../schema/blocks/form";
import { usePageMutation } from "../../procedures/mutation";
import { useTranslation } from "react-i18next";

type FormProps = {
  blockProps: FormType;
  blockPath: BlockPathType;
};

export const Form = ({ blockProps, blockPath }: FormProps) => {
  const { mutation, trigger } = blockProps;
  const { mutate, isPending, error, isSuccess } = usePageMutation();

  const { t } = useTranslation();
  return (
    <form
      id={mutation.id}
      name={mutation.id}
      onSubmit={(e) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const formValues = Object.fromEntries(formData.entries());
        mutate({
          mutation,
          options: { variables: { form: { [mutation.id]: formValues } } },
        });
      }}
    >
      {error && (
        <Alert variant="error" className="mb-4">
          {error.message}
        </Alert>
      )}

      {isSuccess ? (
        <Alert variant="success" className="mb-4">
          {t("direktivPage.page.blocks.form.success")}
        </Alert>
      ) : (
        <>
          <BlockList path={blockPath}>
            {blockProps.blocks.map((block, index) => (
              <Block
                key={index}
                block={block}
                blockPath={[...blockPath, index]}
              />
            ))}
          </BlockList>
          <div className="mt-4 flex justify-end">
            <Button
              disabled={isPending}
              loading={isPending}
              blockProps={trigger}
              as="button"
            />
          </div>
        </>
      )}
    </form>
  );
};
