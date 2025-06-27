import { FieldValues, UseFormReturn } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";

import { AllBlocksType } from "../../schema/blocks";
import { BlockEditFormProps } from "..";
import { DialogFooter } from "./Footer";
import { DialogHeader } from "./Header";
import { ReactNode } from "react";

interface FormWrapperProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  onSubmit: (data: T) => void;
  action: BlockEditFormProps<T>["action"];
  path: BlockEditFormProps<T>["path"];
  blockType: AllBlocksType["type"];
  children: ReactNode;
}

const formId = "block-editor-form";

export const FormWrapper = <T extends FieldValues>({
  form,
  onSubmit,
  action,
  path,
  blockType,
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
      <DialogHeader action={action} path={path} type={blockType} />
      {errors && <FormErrors errors={errors as errorsType} />}
      {children}
      <DialogFooter formId={formId} />
    </form>
  );
};
