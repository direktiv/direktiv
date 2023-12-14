import { EndpointFormSchema, EndpointFormSchemaType } from "../utils";
import { UseFormReturn, useForm } from "react-hook-form";

import { FC } from "react";
import Input from "~/design/Input";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  endpointConfig?: EndpointFormSchemaType;
  children: ({
    formControls,
  }: {
    formControls: UseFormReturn<EndpointFormSchemaType>;
    formMarkup: JSX.Element;
  }) => JSX.Element;
};

export const Form: FC<FormProps> = ({ endpointConfig, children }) => {
  const formControls = useForm<EndpointFormSchemaType>({
    resolver: zodResolver(EndpointFormSchema),
    defaultValues: {
      ...endpointConfig,
    },
  });

  const { register, watch } = formControls;
  return children({
    formControls,
    formMarkup: (
      <>
        <Input {...register("path")} />
        {watch("direktiv_api")}
        <hr />
        {watch("path")}
      </>
    ),
  });
};
