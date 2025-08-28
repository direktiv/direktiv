import { Block, BlockPathType } from "..";

import Alert from "~/design/Alert";
import { BlockList } from "../utils/BlockList";
import { Button } from "../Button";
import { FormType } from "../../../schema/blocks/form";
import { StopPropagation } from "~/components/StopPropagation";
import { createLocalFormVariables } from "../formPrimitives/utils";
import { usePageMutation } from "../../procedures/mutation";
import { useState } from "react";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

type FormProps = {
  blockProps: FormType;
  blockPath: BlockPathType;
};

export const Form = ({ blockProps, blockPath }: FormProps) => {
  const { mutation, trigger } = blockProps;
  const { t } = useTranslation();
  const { toast } = useToast();
  const [missingFields, setMissingFields] = useState<string[]>([]);

  const { mutate, isPending, isSuccess } = usePageMutation({
    onError: (error) => {
      toast({
        title: t("direktivPage.page.blocks.form.error"),
        description: error.message,
        variant: "error",
        duration: Infinity,
      });
    },
  });
  return (
    <form
      id={mutation.id}
      name={mutation.id}
      onSubmit={(formEvent) => {
        formEvent.preventDefault();
        const { formVariables, missingRequiredFields } =
          createLocalFormVariables(formEvent);

        if (missingRequiredFields.length > 0) {
          setMissingFields(missingRequiredFields);
          return;
        }
        mutate({ mutation, formVariables });
      }}
    >
      {isSuccess ? (
        <Alert variant="success" className="mb-4">
          {t("direktivPage.page.blocks.form.success")}
        </Alert>
      ) : (
        <>
          {missingFields.length > 0 && (
            <Alert variant="error" className="mb-4">
              {t("direktivPage.page.blocks.form.incompleteForm")}{" "}
              {missingFields.join(", ")}
            </Alert>
          )}
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
            <StopPropagation>
              <Button
                disabled={isPending}
                loading={isPending}
                blockProps={trigger}
                as="button"
              />
            </StopPropagation>
          </div>
        </>
      )}
    </form>
  );
};
