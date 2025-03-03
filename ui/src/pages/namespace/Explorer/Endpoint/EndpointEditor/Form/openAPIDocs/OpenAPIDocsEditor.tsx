import {
  MethodsSchema,
  MethodsSchemaType,
  RouteMethod,
  routeMethods,
} from "~/api/gateway/schema";
import { jsonToYaml, yamlToJsonOrNull } from "~/pages/namespace/Explorer/utils";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { FC } from "react";
import FormErrors from "~/components/FormErrors";
import { useForm } from "react-hook-form";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type OpenAPIDocsFormProps = {
  defaultValue: MethodsSchemaType;
  onSubmit: (value: MethodsSchemaType) => void;
  readOnly?: boolean;
};

const FormSchema = z.object({
  /**
   * Passthrough is required here to detect if the user adds some additional unallowed
   * keys, that we will then restrict with an error message in the schemas refine function.
   */
  editor: MethodsSchema.passthrough(),
});

type FormSchemaType = z.infer<typeof FormSchema>;

export const OpenAPIDocsEditor: FC<OpenAPIDocsFormProps> = ({
  readOnly,
  defaultValue,
  onSubmit,
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const {
    handleSubmit,
    getValues,
    setValue,
    formState: { errors },
  } = useForm<FormSchemaType>({
    resolver: zodResolver(
      FormSchema.refine(
        (data) => {
          const containsUnsupportedMethod = Object.keys(data.editor).some(
            (method) => !routeMethods.has(method as RouteMethod)
          );
          if (containsUnsupportedMethod) return false;
          return true;
        },
        {
          message: t(
            "pages.explorer.endpoint.editor.form.docs.modal.unsupportedMethods",
            { methods: Array.from(routeMethods).join(", ") }
          ),
        }
      )
    ),
    defaultValues: {
      editor: {
        ...defaultValue,
      },
    },
  });

  const onEditorSubmit = (configuration: FormSchemaType) => {
    onSubmit(configuration.editor);
  };

  return (
    // TODO: this is a hack to get the form to work with the modal
    // disabled
    <form onSubmit={handleSubmit(onEditorSubmit)}>
      <Card className="h-96 w-full p-4" noShadow background="weight-1">
        <FormErrors errors={errors} className="mb-5" />
        <Editor
          defaultValue={jsonToYaml(getValues("editor"))}
          onChange={(newDocs) => {
            if (newDocs !== undefined) {
              const docsAsJson = yamlToJsonOrNull(newDocs) ?? {};
              setValue("editor", docsAsJson);
            }
          }}
          theme={theme ?? undefined}
          options={{ readOnly }}
        />
      </Card>
    </form>
  );
};
