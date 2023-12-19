import { Controller, useForm } from "react-hook-form";
import { FC, FormEvent } from "react";
import FormErrors, { errorsType } from "~/componentsNext/FormErrors";
import {
  JsInboundFormSchema,
  JsInboundFormSchemaType,
} from "../../../schema/plugins/inbound/jsInbound";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DialogFooter } from "~/design/Dialog";
import Editor from "~/design/Editor";
import { useTheme } from "~/util/store/theme";
import { zodResolver } from "@hookform/resolvers/zod";

type OptionalConfig = Partial<JsInboundFormSchemaType["configuration"]>;

type FormProps = {
  defaultConfig?: OptionalConfig;
  onSubmit: (data: JsInboundFormSchemaType) => void;
};

export const JsInboundForm: FC<FormProps> = ({ defaultConfig, onSubmit }) => {
  const {
    handleSubmit,
    formState: { errors },
    control,
  } = useForm<JsInboundFormSchemaType>({
    resolver: zodResolver(JsInboundFormSchema),
    defaultValues: {
      type: "js-inbound",
      configuration: {
        ...defaultConfig,
      },
    },
  });

  const submitForm = (e: FormEvent<HTMLFormElement>) => {
    e.stopPropagation(); // prevent the parent form from submitting
    handleSubmit(onSubmit)(e);
  };

  const theme = useTheme();

  return (
    <form onSubmit={submitForm}>
      {errors?.configuration && (
        <FormErrors
          errors={errors?.configuration as errorsType}
          className="mb-5"
        />
      )}

      <div className="my-3 flex flex-col gap-y-5">
        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] overflow-hidden text-right text-sm">
            script
          </label>
          <Card className="h-[200px] w-full p-5" background="weight-1" noShadow>
            <Controller
              control={control}
              name="configuration.script"
              render={({ field }) => (
                <Editor
                  theme={theme ?? undefined}
                  language="plaintext"
                  value={field.value}
                  onChange={field.onChange}
                />
              )}
            />
          </Card>
        </fieldset>
      </div>
      <DialogFooter>
        <Button type="submit">Save</Button>
      </DialogFooter>
    </form>
  );
};
