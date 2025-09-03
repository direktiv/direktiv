import { Text as TextSchema, TextType } from "../schema/blocks/text";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import { SmartInput } from "./components/SmartInput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
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
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.text.contentLabel")}
        htmlFor="content"
      >
        <SmartInput
          value={form.watch("content")}
          onChange={(content) => form.setValue("content", content)}
          id="content"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.text.contentPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};
