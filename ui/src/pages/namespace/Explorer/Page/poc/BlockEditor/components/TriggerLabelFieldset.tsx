import { DialogType } from "../../schema/blocks/dialog";
import { Fieldset } from "~/components/Form/Fieldset";
import { SmartInput } from "./SmartInput";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type TriggerLabelFieldsetProps = {
  form: UseFormReturn<DialogType>;
};

export const TriggerLabelFieldset = ({ form }: TriggerLabelFieldsetProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset
      label={t("direktivPage.blockEditor.blockForms.dialog.triggerLabelLabel")}
      htmlFor="label"
    >
      <SmartInput
        value={form.watch("trigger.label")}
        onUpdate={(value) =>
          form.setValue("trigger.label", value, { shouldDirty: true })
        }
        id="label"
        placeholder={t(
          "direktivPage.blockEditor.blockForms.dialog.triggerLabelPlaceholder"
        )}
      />
    </Fieldset>
  );
};
