import { FieldValues, UseFormReturn } from "react-hook-form";
import FormErrors, { errorsType } from "~/components/FormErrors";
import { ReactNode, useEffect } from "react";

import { BlockEditFormProps } from "..";
import { BlockType } from "../../schema/blocks";
import { Footer } from "./Footer";
import { Header } from "./Header";
import { NavigationBlocker } from "~/components/NavigationBlocker";
import { usePageEditorPanel } from "../EditorPanelProvider";

interface FormWrapperProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  description: string;
  onSubmit: (data: T) => void;
  onCancel: () => void;
  action: BlockEditFormProps<T>["action"];
  path: BlockEditFormProps<T>["path"];
  block: BlockType;
  children: ReactNode;
}

const formId = "block-editor-form";

export const FormWrapper = <T extends FieldValues>({
  form,
  onSubmit,
  action,
  path,
  block,
  description,
  children,
  onCancel,
}: FormWrapperProps<T>) => {
  const {
    handleSubmit,
    formState: { errors, isDirty },
  } = form;
  const { setDirty } = usePageEditorPanel();

  useEffect(() => setDirty(isDirty), [isDirty, setDirty]);

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      id={formId}
      className="flex flex-col"
    >
      {isDirty && <NavigationBlocker />}
      <div className="flex flex-col gap-4 px-1">
        <div className="p-4">
          <Header action={action} path={path} block={block} />
          <div className="text-gray-10 dark:text-gray-10">{description}</div>
          {errors && <FormErrors errors={errors as errorsType} />}
          {children}
        </div>
      </div>
      <div className="shrink-0 px-4">
        <Footer formId={formId} onCancel={onCancel} />
      </div>
    </form>
  );
};
