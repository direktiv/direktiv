import {
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
  useWatch,
} from "react-hook-form";
import { ServiceFormSchema, ServiceFormSchemaType } from "./schema";

import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: ServiceFormSchemaType;
  children: (args: {
    formControls: UseFormReturn<ServiceFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<ServiceFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const formControls = useForm<ServiceFormSchemaType>({
    resolver: zodResolver(ServiceFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const values = useWatch({
    control: formControls.control,
  });

  const { register, control } = formControls;

  return children({
    formControls,
    values,
    formMarkup: (
      <div className="flex flex-col gap-8">
        <div className="flex gap-3">
          <Fieldset
            label={t("pages.explorer.service.editor.form.image")}
            htmlFor="image"
            className="grow"
          >
            <Input {...register("image")} id="image" />
          </Fieldset>
        </div>
      </div>
    ),
  });
};
