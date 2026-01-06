import {
  FormTextarea as FormTextareaSchema,
  FormTextareaType,
} from "../../schema/blocks/form/textarea";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import { SmartInput } from "../components/SmartInput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TextareaProps = BlockEditFormProps<FormTextareaType>;

export const Textarea = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: TextareaProps) => {
  const { t } = useTranslation();
  const form = useForm<FormTextareaType>({
    resolver: zodResolver(FormTextareaSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.textarea.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <BaseForm form={form} />
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.label"
        )}
        htmlFor="defaultValue"
      >
        <SmartInput
          value={form.watch("defaultValue")}
          onUpdate={(value) =>
            form.setValue("defaultValue", value, { shouldDirty: true })
          }
          id="defaultValue"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.formPrimitives.defaultValue.placeholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};
