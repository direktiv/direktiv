import {
  InstantResponseFormSchema,
  InstantResponseFormSchemaType,
} from "../../../schema/plugins/target/InstantResponse";

import { FC } from "react";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type FormProps = {
  defaultConfig?: InstantResponseFormSchemaType;
};

export const InstantResponseForm: FC<FormProps> = ({ defaultConfig }) => {
  const { register } = useForm<InstantResponseFormSchemaType>({
    resolver: zodResolver(InstantResponseFormSchema),
    defaultValues: { ...defaultConfig },
  });

  return (
    <div className="flex flex-col gap-y-5">
      <fieldset className="flex items-center gap-5">
        <label className="w-[250px] overflow-hidden text-right text-sm">
          content_type
        </label>
        <Input
          {...register("configuration.content_type")}
          placeholder="application/json"
        />
      </fieldset>
      <fieldset className="flex items-center gap-5">
        <label className="w-[250px] overflow-hidden text-right text-sm">
          status_code
        </label>
        <Input
          {...register("configuration.status_code", {
            valueAsNumber: true,
          })}
          type="number"
          placeholder="200"
        />
      </fieldset>
      <fieldset className="flex items-center gap-5">
        <label className="w-[250px] overflow-hidden text-right text-sm">
          status_message
        </label>
        <Textarea {...register("configuration.status_message")} />
      </fieldset>
    </div>
  );
};
