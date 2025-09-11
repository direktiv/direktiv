import { Dialog as DialogSchema, DialogType } from "../schema/blocks/dialog";

import { BlockEditFormProps } from ".";
import { Fieldset } from "~/components/Form/Fieldset";
import { FormWrapper } from "./components/FormWrapper";
import { SmartInput } from "./components/SmartInput";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type DialogFormProps = BlockEditFormProps<DialogType>;

export const Dialog = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
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
      block={propBlock}
      action={action}
      path={path}
      onSubmit={onSubmit}
      onCancel={onCancel}
    >
      <Fieldset
        label={t(
          "direktivPage.blockEditor.blockForms.dialog.triggerLabelLabel"
        )}
        htmlFor="label"
      >
        <SmartInput
          value={form.watch("trigger.label")}
          onUpdate={(value) => form.setValue("trigger.label", value)}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.dialog.triggerLabelPlaceholder"
          )}
        />
      </Fieldset>
    </FormWrapper>
  );
};
