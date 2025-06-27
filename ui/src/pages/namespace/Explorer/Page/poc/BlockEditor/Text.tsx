import { Text as TextSchema, TextType } from "../schema/blocks/text";

import { BlockEditFormProps } from ".";
import { FormWrapper } from "./components/FormWrapper";
import { Textarea } from "~/design/TextArea";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TextBlockEditFormProps) => {
  const form = useForm<TextType>({
    resolver: zodResolver(TextSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      form={form}
      onSubmit={onSubmit}
      action={action}
      path={path}
      blockType={propBlock.type}
    >
      <Textarea {...form.register("content")} />
    </FormWrapper>
  );
};
