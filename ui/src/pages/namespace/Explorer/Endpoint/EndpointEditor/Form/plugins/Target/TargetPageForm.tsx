import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/components/FormErrors";
import {
  TargetPageFormSchema,
  TargetPageFormSchemaType,
} from "../../../schema/plugins/target/targetPage";

import { Fieldset } from "~/components/Form/Fieldset";
import FilePicker from "~/components/FilePicker";
import { PluginWrapper } from "../components/PluginSelector";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<TargetPageFormSchemaType["configuration"]>;

type FormProps = {
  formId: string;
  defaultConfig?: OptionalConfig;
  onSubmit: (data: TargetPageFormSchemaType) => void;
};

export const TargetPageForm: FC<FormProps> = ({
  defaultConfig,
  onSubmit,
  formId,
}) => {
  const { t } = useTranslation();
  const {
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<TargetPageFormSchemaType>({
    resolver: zodResolver(TargetPageFormSchema),
    defaultValues: {
      type: "target-page",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  return (
    <form onSubmit={submitForm} id={formId}>
      <PluginWrapper>
        {errors?.configuration && (
          <FormErrors
            errors={errors?.configuration as errorsType}
            className="mb-5"
          />
        )}
        <Fieldset
          label={t(
            "pages.explorer.endpoint.editor.form.plugins.target.targetPage.file"
          )}
        >
          <Controller
            control={control}
            name="configuration.file"
            render={({ field }) => (
              <FilePicker
                onChange={field.onChange}
                defaultPath={field.value}
                selectable={(file) => file.type === "page"}
              />
            )}
          />
        </Fieldset>
      </PluginWrapper>
    </form>
  );
};
