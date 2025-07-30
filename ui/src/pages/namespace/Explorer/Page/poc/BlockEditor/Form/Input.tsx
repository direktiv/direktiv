import {
  FormInput as FormInputSchema,
  FormInputType,
} from "../../schema/blocks/form/input";
import { FormProvider, useForm } from "react-hook-form";

import { BaseForm } from "./BaseForm";
import { BlockEditFormProps } from "..";
import { FormWrapper } from "../components/FormWrapper";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type InputProps = BlockEditFormProps<FormInputType>;

export const Input = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: InputProps) => {
  const { t } = useTranslation();
  const form = useForm<FormInputType>({
    resolver: zodResolver(FormInputSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t(
        "direktivPage.blockEditor.blockForms.formPrimitives.input.description"
      )}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <FormProvider {...form}>
        <BaseForm />
      </FormProvider>
    </FormWrapper>
  );
};
