import {
  FormDateInputType,
  FormDateInput as FormInputSchema,
} from "../../schema/blocks/form/dateInput";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "../components/FormWrapper";
import InputDesignComponent from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type DateInputProps = BlockEditFormProps<FormDateInputType>;

export const DateInput = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: DateInputProps) => {
  const { t } = useTranslation();
  const form = useForm<FormDateInputType>({
    resolver: zodResolver(FormInputSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.dateInput.description"
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
        <InputDesignComponent
          {...form.register("defaultValue")}
          id="defaultValue"
        />
      </Fieldset>
    </FormWrapper>
  );
};
