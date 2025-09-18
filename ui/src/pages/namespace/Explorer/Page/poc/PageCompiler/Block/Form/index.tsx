import { Block, BlockPathType } from "..";
import {
  FormValidationContextProvider,
  useFormValidation,
} from "./FormValidationContext";
import {
  createLocalFormVariables,
  extractFormKeys,
} from "../formPrimitives/utils";
import { useCallback, useEffect, useRef } from "react";

import Alert from "~/design/Alert";
import { BlockList } from "../utils/BlockList";
import { Button } from "../Button";
import { FormType } from "../../../schema/blocks/form";
import { LocalVariables } from "../../primitives/Variable/VariableContext";
import { StopPropagation } from "~/components/StopPropagation";
import { usePageMutation } from "../../procedures/mutation";
import { useToast } from "~/design/Toast";
import { useTranslation } from "react-i18next";

type FormProps = {
  blockProps: FormType;
  blockPath: BlockPathType;
  register?: (localVariables: LocalVariables) => void;
};

const FormWithContext = ({ blockProps, blockPath, register }: FormProps) => {
  const { mutation, trigger } = blockProps;
  const { t } = useTranslation();
  const { toast } = useToast();
  const { missingFields, setMissingFields } = useFormValidation();
  const ref = useRef<HTMLFormElement>(null);

  const missingFieldsNote =
    missingFields.length > 0 &&
    t("direktivPage.page.blocks.form.incompleteForm", {
      fields: missingFields.join(", "),
    });

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

  const registerForm = useCallback(
    (form: HTMLFormElement | null) => {
      if (!register || !form) return;
      const localVariables = extractFormKeys(form.elements);
      register({ this: localVariables });
    },
    [register]
  );

  return (
    <form
      ref={registerForm}
      onSubmit={(formEvent) => {
        setMissingFields([]);
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
          {missingFieldsNote && (
            <Alert variant="error" className="mb-4">
              {missingFieldsNote}
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
          <div className="mt-4 flex items-center justify-end gap-3">
            {missingFieldsNote && (
              <span className="text-sm text-danger-11 dark:text-danger-dark-11">
                {missingFieldsNote}
              </span>
            )}
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

export const Form = (props: FormProps) => (
  <FormValidationContextProvider>
    <FormWithContext {...props} />
  </FormValidationContextProvider>
);
