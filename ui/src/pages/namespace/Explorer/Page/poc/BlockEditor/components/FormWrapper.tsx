import { FieldValues, UseFormReturn } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { DialogFooter } from "./Footer";
import { DialogHeader } from "./Header";
import { ReactNode } from "react";

interface FormWrapperProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  onSubmit: (data: T) => void;
  children: ReactNode;
}

const formId = "block-editor-form";

export const FormWrapper = <T extends FieldValues>({
  form,
  onSubmit,
  children,
}: FormWrapperProps<T>) => {
  const {
    handleSubmit,
    formState: { errors },
  } = form;

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      id={formId}
      className="flex flex-col gap-3"
    >
      <DialogHeader />
      {errors && <FormErrors errors={errors as errorsType} />}
      {children}
      <DialogFooter formId={formId} />
    </form>
  );
};
