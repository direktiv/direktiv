import { Text as TextSchema, TextType } from "../schema/blocks/text";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import { Textarea } from "~/design/TextArea";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TextBlockEditFormProps) => {
  const { t } = useTranslation();
  const form = useForm<TextType>({
    resolver: zodResolver(TextSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.text.description")}
      form={form}
      block={propBlock}
      action={action}
      path={path}
      blockType={propBlock.type}
      onSubmit={onSubmit}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.text.contentLabel")}
        htmlFor="content"
      >
        <Textarea
          {...form.register("content")}
          id="content"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.text.contentPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};
