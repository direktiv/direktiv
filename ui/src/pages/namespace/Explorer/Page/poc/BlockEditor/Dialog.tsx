import { Dialog as DialogSchema, DialogType } from "../schema/blocks/dialog";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type DialogFormProps = BlockEditFormProps<DialogType>;

export const Dialog = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: DialogFormProps) => {
  const { t } = useTranslation();
  const form = useForm<DialogType>({
    resolver: zodResolver(DialogSchema),
    defaultValues: propBlock,
  });

  return (
    <FormWrapper
      description={t("direktivPage.blockEditor.blockForms.dialog.description")}
      form={form}
      onSubmit={onSubmit}
      action={action}
      path={path}
      blockType={propBlock.type}
    >
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.dialog.triggerLabelLabel"
        )}
        htmlFor="label"
      >
        <Input
          {...form.register("trigger.label")}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.dialog.triggerLabelPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};
