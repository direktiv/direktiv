import {
  FormCheckbox as FormCheckboxSchema,
  FormCheckboxType,
} from "../../schema/blocks/form/checkbox";
import { UseFormReturn, useForm } from "react-hook-form";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Checkbox as CheckboxDesignComponent } from "~/design/Checkbox";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormBaseType } from "../../schema/blocks/form/utils";
import { FormWrapper } from "../components/FormWrapper";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type CheckboxProps = BlockEditFormProps<FormCheckboxType>;

export const Checkbox = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: CheckboxProps) => {
  const { t } = useTranslation();
  const form = useForm<FormCheckboxType>({
    resolver: zodResolver(FormCheckboxSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.checkbox.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <BaseForm form={form as unknown as UseFormReturn<FormBaseType>} />
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.formPrimitives.checkbox.defaultValueLabel"
        )}
        htmlFor="defaultValue"
        horizontal
      >
        <CheckboxDesignComponent
          defaultChecked={form.getValues("defaultValue")}
          onCheckedChange={(value) => {
            if (typeof value === "boolean") {
              form.setValue("defaultValue", value);
            }
          }}
          id="defaultValue"
        />
      </Fieldset>
    </FormWrapper>
  );
};
