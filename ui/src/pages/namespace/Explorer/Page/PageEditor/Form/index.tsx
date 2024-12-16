import {
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
} from "react-hook-form";
import { PageFormSchema, PageFormSchemaType } from "../schema";

import { FC } from "react";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig: PageFormSchemaType;
  onSave: (value: PageFormSchemaType) => void;
  children: (args: {
    formControls: UseFormReturn<PageFormSchemaType>;
    formMarkup: JSX.Element;
    values: DeepPartialSkipArrayKey<PageFormSchemaType>;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ defaultConfig, children }) => {
  const { t } = useTranslation();
  const formControls = useForm<PageFormSchemaType>({
    resolver: zodResolver(PageFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  // const values = useSortedValues(formControls.control);
  const values = defaultConfig;
  const { register } = formControls;

  return children({
    formControls,
    values,
    formMarkup: (
      <div className="flex flex-col gap-8">
        <div className="flex gap-3">
          <Fieldset
            label={t("pages.explorer.endpoint.editor.form.path")}
            htmlFor="path"
            className="grow"
          >
            <Input {...register("path")} id="path" />
          </Fieldset>
        </div>
      </div>
    ),
  });
};
