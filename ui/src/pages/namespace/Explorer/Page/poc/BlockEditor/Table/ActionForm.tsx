import FormErrors, { errorsType } from "~/components/FormErrors";
import { TriggerBlocks, TriggerBlocksType } from "../../schema/blocks";

import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type ActionFormProps = {
  defaultValues?: TriggerBlocksType;
  formId: string;
  onSubmit: (data: TriggerBlocksType) => void;
};

export const ActionForm = ({
  defaultValues,
  formId,
  onSubmit,
}: ActionFormProps) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    register,
    formState: { errors },
  } = useForm<TriggerBlocksType>({
    resolver: zodResolver(TriggerBlocks),
    defaultValues: {
      type: "button",
      label: "",
      ...defaultValues,
    },
  });

  const onFormSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.stopPropagation();
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={onFormSubmit} id={formId}>
      {errors && <FormErrors errors={errors as errorsType} className="mb-5" />}
      <Fieldset
        label={t("direktivPage.blockEditor.blockForms.table.action.labelLabel")}
        htmlFor="label"
      >
        <Input
          {...register("label")}
          id="label"
          placeholder={t(
            "direktivPage.blockEditor.blockForms.table.action.labelPlaceholder"
          )}
        />
      </Fieldset>
    </form>
  );
};
