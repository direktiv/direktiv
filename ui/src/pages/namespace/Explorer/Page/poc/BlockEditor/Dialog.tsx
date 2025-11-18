import { Dialog as DialogSchema, DialogType } from "../schema/blocks/dialog";

import { BlockEditFormProps } from ".";
import { FormWrapper } from "./components/FormWrapper";
import { TriggerLabelFieldset } from "./components/TriggerLabelFieldset";
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
      <TriggerLabelFieldset form={form} />
    </FormWrapper>
  );
};
