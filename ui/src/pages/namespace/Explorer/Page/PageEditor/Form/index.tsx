import {
  DeepPartialSkipArrayKey,
  UseFormReturn,
  useForm,
} from "react-hook-form";
import { PageFormSchema, PageFormSchemaType } from "../schema";

import { FC } from "react";
import Input from "~/design/Input";
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
  const formControls = useForm<PageFormSchemaType>({
    resolver: zodResolver(PageFormSchema),
    defaultValues: {
      ...defaultConfig,
    },
  });

  const values = defaultConfig;
  const { register } = formControls;

  return children({
    formControls,
    values,
    formMarkup: (
      <div className="flex flex-col gap-8">
        <Input className="hidden" {...register("layout")} id="layout" />
      </div>
    ),
  });
};
