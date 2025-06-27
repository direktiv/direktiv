import FormErrors, { errorsType } from "~/components/FormErrors";
import { Text as TextSchema, TextType } from "../schema/blocks/text";

import { BlockEditFormProps } from ".";
import { DialogFooter } from "./components/Footer";
import { DialogHeader } from "./components/Header";
import { Textarea } from "~/design/TextArea";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

const formId = "block-editor-text";

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TextBlockEditFormProps) => {
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<TextType>({
    resolver: zodResolver(TextSchema),
    defaultValues: propBlock,
  });

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      id={formId}
      className="flex flex-col gap-3"
    >
      <DialogHeader action={action} path={path} type={propBlock.type} />
      {errors && <FormErrors errors={errors as errorsType} />}
      <Textarea {...register("content")} />
      <DialogFooter formId={formId} />
    </form>
  );
};
