import { ButtonHTMLAttributes, forwardRef } from "react";
import {
  MethodsSchema,
  MethodsSchemaType,
  RouteMethod,
  routeMethods,
} from "~/api/gateway/schema";
import { jsonToYaml, yamlToJsonOrNull } from "~/pages/namespace/Explorer/utils";

import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import FormErrors from "~/components/FormErrors";
import { useForm } from "react-hook-form";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type HTMLFormProps = ButtonHTMLAttributes<HTMLFormElement>;

type OpenAPIDocsFormProps = {
  defaultValue: MethodsSchemaType;
} & (
  | { onSubmit?: (value: MethodsSchemaType) => void; readOnly?: never }
  | { onSubmit?: never; readOnly?: boolean }
) &
  Omit<HTMLFormProps, "onSubmit" | "defaultValue" | "readOnly">;

const FormSchema = z.object({
  /**
   * Passthrough is required here to detect if the user adds some additional unallowed
   * keys, that we will then restrict with an error message in the schemas refine function.
   */
  editor: MethodsSchema.passthrough(),
});

type FormSchemaType = z.infer<typeof FormSchema>;

export const OpenAPIDocsEditor = forwardRef<
  HTMLFormElement,
  OpenAPIDocsFormProps
>(({ readOnly, defaultValue, onSubmit, ...props }, ref) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const {
    handleSubmit,
    getValues,
    setValue,
    setError,
    clearErrors,
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
    onSubmit?.(configuration.editor);
  };

  const onEditorChange = (newDocs: string | undefined) => {
    if (newDocs === undefined) return;

    // setting everything to an empty string will delete all docs
    if (newDocs.trim() === "") {
      setValue("editor", {});
      return;
    }

    const docsAsJson = yamlToJsonOrNull(newDocs);
    // non empty string that result in null, are errors
    if (docsAsJson === null) {
      setError("root", {
        message: t(
          "pages.explorer.endpoint.editor.form.docs.modal.invalidYaml"
        ),
      });
      return;
    }

    clearErrors();
    setValue("editor", docsAsJson);
  };

  return (
    <form onSubmit={handleSubmit(onEditorSubmit)} ref={ref} {...props}>
      <FormErrors errors={errors} className="mb-5" />
      <Card className="h-96 w-full p-4" noShadow background="weight-1">
        <Editor
          defaultValue={jsonToYaml(getValues("editor"))}
          onChange={onEditorChange}
          theme={theme ?? undefined}
          options={{ readOnly }}
        />
      </Card>
    </form>
  );
});

OpenAPIDocsEditor.displayName = "OpenAPIDocsEditor";
